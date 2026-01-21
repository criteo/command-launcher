package updateConfig

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestReadFromDir(t *testing.T) {
	tmpDir := t.TempDir()

	config := NewUpdateConfig()
	config.PausePackage("test-package", DEFAULT_UPDATE_PAUSE_DURATION)
	err := config.WriteToDir(tmpDir)
	assert.NoError(t, err)

	readConfig, err := ReadFromDir(tmpDir)
	assert.NoError(t, err)
	assert.True(t, readConfig.IsPackagePaused("test-package"))
}

func TestReadFromDir_NotExists(t *testing.T) {
	tmpDir := t.TempDir()

	_, err := ReadFromDir(tmpDir)
	assert.Error(t, err)
}

func TestIsUpdateConfigExists(t *testing.T) {
	tmpDir := t.TempDir()

	exists, err := IsUpdateConfigExists(tmpDir)
	assert.NoError(t, err)
	assert.False(t, exists)

	config := NewUpdateConfig()
	config.PausePackage("test-package", DEFAULT_UPDATE_PAUSE_DURATION)
	err = config.WriteToDir(tmpDir)
	assert.NoError(t, err)

	exists, err = IsUpdateConfigExists(tmpDir)
	assert.NoError(t, err)
	assert.True(t, exists)
}

func TestWriteToDir(t *testing.T) {
	tmpDir := t.TempDir()

	config := NewUpdateConfig()
	config.PausedUntil["test-package"] = time.Date(2025, 11, 14, 12, 0, 0, 0, time.UTC)

	err := config.WriteToDir(tmpDir)
	assert.NoError(t, err)

	pauseFile := filepath.Join(tmpDir, PACKAGE_UPDATE_FILE)
	_, err = os.Stat(pauseFile)
	assert.NoError(t, err)

	readConfig, err := ReadFromDir(tmpDir)
	assert.NoError(t, err)
	assert.Equal(t, config.PausedUntil["test-package"], readConfig.PausedUntil["test-package"])
}

func TestIsPackagePaused(t *testing.T) {
	tests := []struct {
		name        string
		updateAfter time.Time
		expected    bool
	}{
		{
			name:        "Future date - paused",
			updateAfter: time.Now().Add(1 * time.Hour),
			expected:    true,
		},
		{
			name:        "Past date - not paused",
			updateAfter: time.Now().Add(-1 * time.Hour),
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewUpdateConfig()
			config.PausedUntil["test-package"] = tt.updateAfter
			result := config.IsPackagePaused("test-package")
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsPackagePaused_NotExists(t *testing.T) {
	config := NewUpdateConfig()
	assert.False(t, config.IsPackagePaused("non-existent-package"))
}

func TestPausePackage(t *testing.T) {
	config := NewUpdateConfig()
	duration := 48 * time.Hour

	before := time.Now()
	config.PausePackage("test-package", duration)
	after := time.Now()

	expectedMin := before.Add(duration)
	expectedMax := after.Add(duration)

	pauseTime := config.PausedUntil["test-package"]
	assert.True(t, !pauseTime.Before(expectedMin) && !pauseTime.After(expectedMax))
}

func TestRemoveExpiredPauses(t *testing.T) {
	config := NewUpdateConfig()
	config.PausedUntil["expired-package"] = time.Now().Add(-1 * time.Hour)
	config.PausedUntil["active-package"] = time.Now().Add(1 * time.Hour)

	config.RemoveExpiredPauses()

	_, expiredExists := config.PausedUntil["expired-package"]
	_, activeExists := config.PausedUntil["active-package"]

	assert.False(t, expiredExists, "expired package should be removed")
	assert.True(t, activeExists, "active package should remain")
}

func TestMultiplePackages(t *testing.T) {
	tmpDir := t.TempDir()

	config := NewUpdateConfig()
	config.PausePackage("package-a", DEFAULT_UPDATE_PAUSE_DURATION)
	config.PausePackage("package-b", 2*DEFAULT_UPDATE_PAUSE_DURATION)
	err := config.WriteToDir(tmpDir)
	assert.NoError(t, err)

	readConfig, err := ReadFromDir(tmpDir)
	assert.NoError(t, err)
	assert.True(t, readConfig.IsPackagePaused("package-a"))
	assert.True(t, readConfig.IsPackagePaused("package-b"))
	assert.False(t, readConfig.IsPackagePaused("package-c"))
}
