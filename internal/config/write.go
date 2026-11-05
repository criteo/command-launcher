package config

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/gofrs/flock"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// configLockFileName is the lock file, in the app dir, that serializes config
// writes across processes. It is never deleted: a flock is bound to the open
// file descriptor and released by the OS when its holder exits (even on
// SIGKILL), so it cannot go stale, while deleting it could let two processes
// lock two different inodes at the same path.
const configLockFileName = ".config.lock"

const defaultConfigFileMode fs.FileMode = 0644

var (
	lockTimeout    = 5 * time.Second
	lockRetryDelay = 50 * time.Millisecond
	rename         = os.Rename

	// a rename over a file that another process has open fails on Windows;
	// readers only keep it open briefly, so retry with a backoff (~1s total)
	isRetryableRenameErr = isTransientRenameError
	renameAttempts       = 10
	renameRetryDelay     = 10 * time.Millisecond
	renameMaxRetryDelay  = 200 * time.Millisecond
)

// WriteConfig persists the current settings to the active config file.
// Concurrent writers are serialized with a lock file, and the file is replaced
// atomically (temp file + rename) so readers never see a partial file.
func WriteConfig() error {
	return writeConfigFile(viper.GetViper(), configMetadata.File, configLockPath(), false)
}

// SafeWriteConfig is like WriteConfig but leaves an existing file untouched
// (the check is done while holding the lock).
func SafeWriteConfig() error {
	return writeConfigFile(viper.GetViper(), configMetadata.File, configLockPath(), true)
}

func configLockPath() string {
	return filepath.Join(AppDir(), configLockFileName)
}

func writeConfigFile(v *viper.Viper, path string, lockPath string, onlyIfMissing bool) error {
	if path == "" {
		return errors.New("no configuration file to write to")
	}

	lock, err := lockConfig(lockPath)
	if err != nil {
		return err
	}
	defer lock.Unlock()

	target := resolveSymlink(path)
	removeOrphanTempFiles(target)

	if onlyIfMissing {
		if _, err := os.Stat(target); err == nil {
			return nil
		}
	}

	return replaceFile(v, target)
}

// resolveSymlink returns the file a symlinked config points to, so that the
// rename replaces that file instead of turning the link into a regular file
// (dotfile managers commonly symlink the config).
func resolveSymlink(path string) string {
	fi, err := os.Lstat(path)
	if err != nil || fi.Mode()&os.ModeSymlink == 0 {
		return path
	}
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		log.Debugf("cannot resolve configuration symlink %s: %v", path, err)
		return path
	}
	return resolved
}

func lockConfig(lockPath string) (*flock.Flock, error) {
	if err := os.MkdirAll(filepath.Dir(lockPath), 0755); err != nil {
		return nil, fmt.Errorf("cannot create config lock directory: %w", err)
	}

	lock := flock.New(lockPath)
	ctx, cancel := context.WithTimeout(context.Background(), lockTimeout)
	defer cancel()

	locked, err := lock.TryLockContext(ctx, lockRetryDelay)
	if err == nil && !locked {
		err = errors.New("lock not acquired")
	}
	if err != nil {
		lock.Close()
		return nil, fmt.Errorf("cannot lock configuration file (%s) within %s: %w", lockPath, lockTimeout, err)
	}
	return lock, nil
}

func tempFilePattern(target string) string {
	return "." + filepath.Base(target) + ".*.tmp"
}

// removeOrphanTempFiles deletes temp files left behind by a writer that was
// killed before renaming. It must be called with the lock held: no other
// writer can be running, so every matching file is an orphan.
func removeOrphanTempFiles(target string) {
	matches, err := filepath.Glob(filepath.Join(filepath.Dir(target), tempFilePattern(target)))
	if err != nil {
		return
	}
	for _, m := range matches {
		if err := os.Remove(m); err != nil {
			log.Debugf("cannot remove orphan temporary configuration file %s: %v", m, err)
		}
	}
}

// replaceFile writes the configuration to a temp file in the target's
// directory (rename is only atomic within a filesystem), flushes it to disk
// and renames it over the target. The target is left untouched on any error.
// The existing file's mode (and owner, when running as root) is kept.
func replaceFile(v *viper.Viper, target string) (err error) {
	mode := defaultConfigFileMode
	existing, statErr := os.Stat(target)
	if statErr == nil {
		mode = existing.Mode().Perm()
	}

	tmp, err := os.CreateTemp(filepath.Dir(target), tempFilePattern(target))
	if errors.Is(err, fs.ErrPermission) {
		// the directory is read-only but the file itself may be writable
		log.Debugf("cannot create a temporary file next to %s (%v), writing it in place", target, err)
		return writeInPlace(v, target, mode)
	}
	if err != nil {
		return fmt.Errorf("cannot create temporary configuration file: %w", err)
	}
	tmpName := tmp.Name()
	defer func() {
		if err != nil {
			tmp.Close()
			os.Remove(tmpName)
		}
	}()

	if err = v.WriteConfigTo(tmp); err != nil {
		return fmt.Errorf("cannot write configuration: %w", err)
	}
	// CreateTemp creates the file with mode 0600
	if err = tmp.Chmod(mode); err != nil {
		return fmt.Errorf("cannot set configuration file mode: %w", err)
	}
	if existing != nil {
		if err = chownLike(tmp, existing); err != nil {
			return fmt.Errorf("cannot set configuration file owner: %w", err)
		}
	}
	if err = tmp.Sync(); err != nil {
		return fmt.Errorf("cannot flush configuration: %w", err)
	}
	if err = tmp.Close(); err != nil {
		return fmt.Errorf("cannot close temporary configuration file: %w", err)
	}
	if err = renameWithRetry(tmpName, target); err != nil {
		return fmt.Errorf("cannot replace configuration file %s: %w", target, err)
	}
	// persist the rename itself, otherwise a crash can bring back the old file
	// or, on some filesystems, an empty one
	if syncErr := syncDir(filepath.Dir(target)); syncErr != nil {
		log.Debugf("cannot sync configuration directory: %v", syncErr)
	}
	return nil
}

func renameWithRetry(src, dst string) error {
	delay := renameRetryDelay
	for attempt := 1; ; attempt++ {
		err := rename(src, dst)
		if err == nil || attempt >= renameAttempts || !isRetryableRenameErr(err) {
			return err
		}
		log.Debugf("rename %s failed (attempt %d/%d), retrying in %s: %v", dst, attempt, renameAttempts, delay, err)
		time.Sleep(delay)
		delay = min(2*delay, renameMaxRetryDelay)
	}
}

// writeInPlace is the fallback when no temp file can be created next to the
// target: it truncates and rewrites the file like viper does. Concurrent
// writers are still serialized by the lock, but a reader may see a partial file.
func writeInPlace(v *viper.Viper, target string, mode fs.FileMode) error {
	f, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("cannot open configuration file: %w", err)
	}
	defer f.Close()

	if err := v.WriteConfigTo(f); err != nil {
		return fmt.Errorf("cannot write configuration: %w", err)
	}
	if err := f.Sync(); err != nil {
		return fmt.Errorf("cannot flush configuration: %w", err)
	}
	return f.Close()
}
