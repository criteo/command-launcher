package updater

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/criteo/command-launcher/internal/config"
	"github.com/criteo/command-launcher/internal/user"
	"github.com/stretchr/testify/assert"
)

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
		Policy:           config.SelfUpdatePolicyOnlyNewer,
	}

	updateChan := updater.checkSelfUpdate()

	select {
	case shouldUpdate := <-updateChan:
		assert.False(t, shouldUpdate, "User outside partition range should not get update")
	case <-time.After(6 * time.Second):
		t.Fatal("Timeout waiting for update check result")
	}
}

func TestSelfUpdaterCheckUpdateWithDifferentPolicies(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "self-updater-policy-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a test version file
	versionFile := filepath.Join(tempDir, "version.yaml")
	versionContent := `
version: 1.14.0-7
releaseNotes: "Test release notes"
startPartition: 0
endPartition: 255
`
	err = os.WriteFile(versionFile, []byte(versionContent), 0644)
	assert.NoError(t, err)

	// Create a mock user
	mockUser := user.User{
		UID:                    "test-user",
		Partition:              5,
		InternalCmdEnabled:     false,
		ExperimentalCmdEnabled: false,
	}

	testCases := []struct {
		currentVersion string
		policy         config.SelfUpdatePolicy
		shouldUpdate   bool
		description    string
	}{
		{
			currentVersion: "1.13.0-0",
			policy:         config.SelfUpdatePolicyExactMatch,
			shouldUpdate:   true,
			description:    "Exact match policy with different version should offer update",
		},
		{
			currentVersion: "1.14.0-7",
			policy:         config.SelfUpdatePolicyExactMatch,
			shouldUpdate:   false,
			description:    "Exact match policy with same version should NOT offer update",
		},
		{
			currentVersion: "1.15.0-0",
			policy:         config.SelfUpdatePolicyExactMatch,
			shouldUpdate:   true,
			description:    "Exact match policy with newer local version should offer update (rollback)",
		},
		{
			currentVersion: "1.13.0-0",
			policy:         config.SelfUpdatePolicyOnlyNewer,
			shouldUpdate:   true,
			description:    "Only newer policy with older local version should offer update",
		},
		{
			currentVersion: "1.14.0-7",
			policy:         config.SelfUpdatePolicyOnlyNewer,
			shouldUpdate:   false,
			description:    "Only newer policy with same version should NOT offer update",
		},
		{
			currentVersion: "1.15.0-0",
			policy:         config.SelfUpdatePolicyOnlyNewer,
			shouldUpdate:   false,
			description:    "Only newer policy with newer local version should NOT offer update",
		},
		{
			currentVersion: "1.15.0-0",
			policy:         "", // Test empty/invalid policy - should default to exact_match
			shouldUpdate:   true,
			description:    "Empty policy should default to exact_match and offer update (rollback)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			updater := &SelfUpdater{
				BinaryName:       "test-binary",
				LatestVersionUrl: "file://" + versionFile,
				CurrentVersion:   tc.currentVersion,
				User:             mockUser,
				Timeout:          5 * time.Second,
				Policy:           tc.policy,
			}

			updateChan := updater.checkSelfUpdate()

			select {
			case shouldUpdate := <-updateChan:
				assert.Equal(t, tc.shouldUpdate, shouldUpdate,
					"Current: %s, Remote: 1.14.0-7, Policy: %s", tc.currentVersion, tc.policy)
			case <-time.After(6 * time.Second):
				t.Fatal("Timeout waiting for update check result")
			}
		})
	}
}
