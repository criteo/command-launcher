package pkg

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsSetupDone_NotExists(t *testing.T) {
	tmpDir := t.TempDir()

	assert.False(t, IsSetupDone(tmpDir))
}

func TestMarkSetupDone(t *testing.T) {
	tmpDir := t.TempDir()

	err := MarkSetupDone(tmpDir, "1.2.3")
	assert.NoError(t, err)

	assert.True(t, IsSetupDone(tmpDir))

	markerFile := filepath.Join(tmpDir, SETUP_MARKER_FILE)
	data, err := os.ReadFile(markerFile)
	assert.NoError(t, err)

	var marker SetupMarker
	err = json.Unmarshal(data, &marker)
	assert.NoError(t, err)
	assert.Equal(t, "1.2.3", marker.PackageVersion)
	assert.False(t, marker.CompletedAt.IsZero())
}

func TestIsSetupDone_Table(t *testing.T) {
	tests := []struct {
		name      string
		setupDone bool
		expected  bool
	}{
		{name: "no marker written", setupDone: false, expected: false},
		{name: "marker written", setupDone: true, expected: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			if tt.setupDone {
				err := MarkSetupDone(tmpDir, "1.0.0")
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expected, IsSetupDone(tmpDir))
		})
	}
}
