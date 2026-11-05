package config

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gofrs/flock"
	"github.com/spf13/viper"
)

// configLockFileName is the lock file, in the app dir, that serializes config
// writes across processes. It is never deleted: a flock is bound to the open
// file descriptor and released by the OS when its holder exits (even on
// SIGKILL), so it cannot go stale, while deleting it could let two processes
// lock two different inodes at the same path.
const configLockFileName = ".config.lock"

var (
	lockTimeout    = 5 * time.Second
	lockRetryDelay = 50 * time.Millisecond
	rename         = os.Rename
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

	if onlyIfMissing {
		if _, err := os.Stat(path); err == nil {
			return nil
		}
	}

	return replaceFile(v, path)
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

// replaceFile writes the configuration to a temp file in the target's
// directory (rename is only atomic within a filesystem), flushes it to disk
// and renames it over the target. The target is left untouched on any error.
func replaceFile(v *viper.Viper, target string) (err error) {
	dir, base := filepath.Split(target)
	tmp, err := os.CreateTemp(dir, "."+base+".*.tmp")
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
	if err = tmp.Sync(); err != nil {
		return fmt.Errorf("cannot flush configuration: %w", err)
	}
	if err = tmp.Close(); err != nil {
		return fmt.Errorf("cannot close temporary configuration file: %w", err)
	}
	if err = rename(tmpName, target); err != nil {
		return fmt.Errorf("cannot replace configuration file %s: %w", target, err)
	}
	return nil
}
