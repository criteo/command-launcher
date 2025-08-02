package updater

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/criteo/command-launcher/internal/user"
	"github.com/stretchr/testify/assert"
)

func TestSelfUpdaterCheckUpdate(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "self-updater-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	testCases := []struct {
		currentVersion string
		remoteVersion  string
		shouldUpdate   bool
		description    string
	}{
		{
			currentVersion: "1.13.0-0",
			remoteVersion:  "1.14.0-7",
			shouldUpdate:   true,
			description:    "Remote version is newer, should offer update",
		},
		{
			currentVersion: "1.14.0-7",
			remoteVersion:  "1.13.0-0",
			shouldUpdate:   false,
			description:    "Remote version is older, should NOT offer update",
		},
		{
			currentVersion: "1.14.0-7",
			remoteVersion:  "1.14.0-7",
			shouldUpdate:   false,
			description:    "Same version, should NOT offer update",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			// Create a test version file
			versionFile := filepath.Join(tempDir, "version.yaml")
			versionContent := fmt.Sprintf(`
version: %s
releaseNotes: "Test release notes"
startPartition: 0
endPartition: 255
`, tc.remoteVersion)

			err := os.WriteFile(versionFile, []byte(versionContent), 0644)
			assert.NoError(t, err)

			// Create a mock user
			mockUser := user.User{
				UID:                    "test-user",
				Partition:              5, // Use a partition within the range 0-255
				InternalCmdEnabled:     false,
				ExperimentalCmdEnabled: false,
			}

			// Create SelfUpdater instance
			updater := &SelfUpdater{
				BinaryName:       "test-binary",
				LatestVersionUrl: "file://" + versionFile,
				CurrentVersion:   tc.currentVersion,
				User:             mockUser,
				Timeout:          5 * time.Second,
			}

			// Test the checkSelfUpdate functionality
			updateChan := updater.checkSelfUpdate()

			// Wait for the result
			select {
			case shouldUpdate := <-updateChan:
				assert.Equal(t, tc.shouldUpdate, shouldUpdate,
					"Current: %s, Remote: %s", tc.currentVersion, tc.remoteVersion)
			case <-time.After(6 * time.Second):
				t.Fatal("Timeout waiting for update check result")
			}
		})
	}
}

func TestSelfUpdaterCheckUpdateWithPartitionFiltering(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "self-updater-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a test version file with partition restrictions
	versionFile := filepath.Join(tempDir, "version.yaml")
	versionContent := `
version: 1.14.0-7
releaseNotes: "Test release notes"
startPartition: 10
endPartition: 20
`

	err = os.WriteFile(versionFile, []byte(versionContent), 0644)
	assert.NoError(t, err)

	// Test with user outside the partition range
	mockUserOutside := user.User{
		UID:                    "test-user-outside",
		Partition:              25, // Use a partition outside the range 10-20
		InternalCmdEnabled:     false,
		ExperimentalCmdEnabled: false,
	}

	updater := &SelfUpdater{
		BinaryName:       "test-binary",
		LatestVersionUrl: "file://" + versionFile,
		CurrentVersion:   "1.13.0-0",
		User:             mockUserOutside,
		Timeout:          5 * time.Second,
	}

	updateChan := updater.checkSelfUpdate()

	select {
	case shouldUpdate := <-updateChan:
		assert.False(t, shouldUpdate, "User outside partition range should not get update")
	case <-time.After(6 * time.Second):
		t.Fatal("Timeout waiting for update check result")
	}
}
