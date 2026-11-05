package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gofrs/flock"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestViper(settings map[string]any) *viper.Viper {
	v := viper.New()
	v.SetConfigType("json")
	for k, val := range settings {
		v.Set(k, val)
	}
	return v
}

func readJSON(t *testing.T, path string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	content := map[string]any{}
	require.NoError(t, json.Unmarshal(data, &content), "invalid JSON: %q", string(data))
	return content
}

func tempFiles(t *testing.T, dir string) []string {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(dir, ".*.tmp"))
	require.NoError(t, err)
	return matches
}

func TestWriteConfig_WritesValidJSON(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config.json")

	err := writeConfigFile(newTestViper(map[string]any{"log_level": "debug"}), cfg, filepath.Join(dir, configLockFileName), false)

	require.NoError(t, err)
	assert.Equal(t, "debug", readJSON(t, cfg)["log_level"])
	assert.Empty(t, tempFiles(t, dir))
}

func TestWriteConfig_OverwritesExisting(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config.json")
	require.NoError(t, os.WriteFile(cfg, []byte(`{"log_level":"fatal"}`), 0644))

	err := writeConfigFile(newTestViper(map[string]any{"log_level": "debug"}), cfg, filepath.Join(dir, configLockFileName), false)

	require.NoError(t, err)
	assert.Equal(t, "debug", readJSON(t, cfg)["log_level"])
}

func TestSafeWriteConfig_DoesNotOverwrite(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config.json")
	require.NoError(t, os.WriteFile(cfg, []byte(`{"log_level":"fatal"}`), 0644))

	err := writeConfigFile(newTestViper(map[string]any{"log_level": "debug"}), cfg, filepath.Join(dir, configLockFileName), true)

	require.NoError(t, err)
	assert.Equal(t, "fatal", readJSON(t, cfg)["log_level"])
}

func TestSafeWriteConfig_CreatesMissing(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config.json")

	err := writeConfigFile(newTestViper(map[string]any{"log_level": "debug"}), cfg, filepath.Join(dir, configLockFileName), true)

	require.NoError(t, err)
	assert.Equal(t, "debug", readJSON(t, cfg)["log_level"])
}

func TestWriteConfig_CreatesLockDir(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config.json")
	lockPath := filepath.Join(dir, "missing", "app", configLockFileName)

	require.NoError(t, writeConfigFile(newTestViper(nil), cfg, lockPath, false))
	assert.FileExists(t, lockPath)
}

func TestWriteConfig_EmptyPath(t *testing.T) {
	err := writeConfigFile(newTestViper(nil), "", filepath.Join(t.TempDir(), configLockFileName), false)
	assert.Error(t, err)
}

func TestWriteConfig_LockTimeout(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config.json")
	lockPath := filepath.Join(dir, configLockFileName)
	require.NoError(t, os.WriteFile(cfg, []byte(`{"log_level":"fatal"}`), 0644))

	holder := flock.New(lockPath)
	locked, err := holder.TryLock()
	require.NoError(t, err)
	require.True(t, locked)
	defer holder.Unlock()

	saved := lockTimeout
	lockTimeout = 100 * time.Millisecond
	defer func() { lockTimeout = saved }()

	start := time.Now()
	err = writeConfigFile(newTestViper(map[string]any{"log_level": "debug"}), cfg, lockPath, false)

	assert.ErrorContains(t, err, "cannot lock configuration file")
	assert.Less(t, time.Since(start), 2*time.Second)
	assert.Equal(t, "fatal", readJSON(t, cfg)["log_level"])
}

func TestWriteConfig_RenameFailureKeepsOriginal(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config.json")
	require.NoError(t, os.WriteFile(cfg, []byte(`{"log_level":"fatal"}`), 0644))

	saved := rename
	rename = func(string, string) error { return errors.New("boom") }
	defer func() { rename = saved }()

	err := writeConfigFile(newTestViper(map[string]any{"log_level": "debug"}), cfg, filepath.Join(dir, configLockFileName), false)

	assert.ErrorContains(t, err, "boom")
	assert.Equal(t, "fatal", readJSON(t, cfg)["log_level"])
	assert.Empty(t, tempFiles(t, dir))
}

// Each writer uses its own viper instance and its own flock (own file
// descriptor), so they contend on the lock like separate processes would.
func TestWriteConfig_ConcurrentWritersAndReaders(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config.json")
	lockPath := filepath.Join(dir, configLockFileName)
	// a large payload makes partial writes much more likely to be observed
	payload := strings.Repeat("x", 64*1024)
	require.NoError(t, writeConfigFile(newTestViper(map[string]any{"writer": -1, "payload": payload}), cfg, lockPath, false))

	const writers, writes = 20, 20
	stop := make(chan struct{})
	readErrs := make(chan error, 1)
	go func() {
		defer close(readErrs)
		for {
			select {
			case <-stop:
				return
			default:
			}
			data, err := os.ReadFile(cfg)
			if err != nil {
				// on Windows a read can race with the rename; it is not a partial file
				continue
			}
			content := map[string]any{}
			if err := json.Unmarshal(data, &content); err != nil {
				readErrs <- fmt.Errorf("reader saw invalid JSON (%d bytes): %w", len(data), err)
				return
			}
		}
	}()

	var wg sync.WaitGroup
	errs := make(chan error, writers*writes)
	for w := 0; w < writers; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			v := newTestViper(map[string]any{"writer": w, "payload": payload})
			for i := 0; i < writes; i++ {
				if err := writeConfigFile(v, cfg, lockPath, false); err != nil {
					errs <- err
				}
			}
		}(w)
	}
	wg.Wait()
	close(stop)
	close(errs)

	for err := range errs {
		assert.NoError(t, err)
	}
	assert.NoError(t, <-readErrs)
	content := readJSON(t, cfg)
	assert.Equal(t, payload, content["payload"])
	assert.Empty(t, tempFiles(t, dir))
}
