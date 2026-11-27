package updater

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/criteo/command-launcher/internal/user"
	"github.com/stretchr/testify/assert"
)

// The latest version file is JSON on some hosts and YAML (one entry of
// release-notes.yaml) on GitHub releases: both must decode the same way.
func TestCheckSelfUpdateReadsJSONAndYAML(t *testing.T) {
	tests := []struct {
		name    string
		payload string
	}{
		{
			name: "json",
			payload: `{
    "version": "2.0.0",
    "releaseNotes": "# Bug Fixes\n  * fix something\n",
    "startPartition": 2,
    "endPartition": 7
}`,
		},
		{
			name: "yaml",
			payload: `version: 2.0.0
releaseNotes: |
  # Bug Fixes
    * fix something
startPartition: 2
endPartition: 7
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "latest")
			assert.Nil(t, os.WriteFile(path, []byte(tt.payload), 0o644))

			u := SelfUpdater{
				LatestVersionUrl: path,
				CurrentVersion:   "1.0.0",
				User:             user.User{Partition: 5},
			}

			assert.True(t, <-u.checkSelfUpdate())
			assert.Equal(t, "2.0.0", u.latestVersion.Version)
			assert.Equal(t, "# Bug Fixes\n  * fix something\n", u.latestVersion.ReleaseNotes)
			assert.Equal(t, uint8(2), u.latestVersion.StartPartition)
			assert.Equal(t, uint8(7), u.latestVersion.EndPartition)
		})
	}
}

func TestCheckSelfUpdateSkipsUsersOutsideThePartition(t *testing.T) {
	path := filepath.Join(t.TempDir(), "latest")
	assert.Nil(t, os.WriteFile(path, []byte("version: 2.0.0\nstartPartition: 0\nendPartition: 1\n"), 0o644))

	u := SelfUpdater{
		LatestVersionUrl: path,
		CurrentVersion:   "1.0.0",
		User:             user.User{Partition: 5},
	}

	assert.False(t, <-u.checkSelfUpdate())
}
