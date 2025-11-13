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

	config := &UpdateConfig{
		Date: time.Now().Add(DEFAULT_UPDATE_LOCK_DURATION),
	}
	err := config.WriteToDir(tmpDir)
	assert.NoError(t, err)

	readConfig, err := ReadFromDir(tmpDir)
	assert.NoError(t, err)
	assert.True(t, readConfig.Date.Equal(config.Date))
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

	config := &UpdateConfig{Date: time.Now()}
	err = config.WriteToDir(tmpDir)
	assert.NoError(t, err)

	exists, err = IsUpdateConfigExists(tmpDir)
	assert.NoError(t, err)
	assert.True(t, exists)
}

func TestWriteToDir(t *testing.T) {
	tmpDir := t.TempDir()

	config := &UpdateConfig{
		Date: time.Date(2025, 11, 14, 12, 0, 0, 0, time.UTC),
	}

	err := config.WriteToDir(tmpDir)
	assert.NoError(t, err)

	lockFile := filepath.Join(tmpDir, PACKAGE_UPDATE_LOCK_FILE)
	_, err = os.Stat(lockFile)
	assert.NoError(t, err)

	readConfig, err := ReadFromDir(tmpDir)
	assert.NoError(t, err)
	assert.True(t, readConfig.Date.Equal(config.Date))
}

func TestIsExpired(t *testing.T) {
	tests := []struct {
		name     string
		date     time.Time
		expected bool
	}{
		{
			name:     "Future date",
			date:     time.Now().Add(1 * time.Hour),
			expected: false,
		},
		{
			name:     "Past date",
			date:     time.Now().Add(-1 * time.Hour),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &UpdateConfig{Date: tt.date}
			result := config.IsExpired()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUpdateAfterDate(t *testing.T) {
	config := &UpdateConfig{}
	duration := 48 * time.Hour

	before := time.Now()
	config.UpdateAfterDate(duration)
	after := time.Now()

	expectedMin := before.Add(duration)
	expectedMax := after.Add(duration)

	assert.True(t, !config.Date.Before(expectedMin) && !config.Date.After(expectedMax))
}
