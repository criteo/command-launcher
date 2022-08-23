package updater

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadLockedPackages(t *testing.T) {
	file, _ := os.Open("assets/lock.json")
	assert.NotNil(t, file)

	cmdUpdater := CmdUpdater{}

	fullPath, _ := filepath.Abs("assets/lock.json")
	lockedPkgs, err := cmdUpdater.LoadLockedPackages(fullPath)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(lockedPkgs))
	assert.Equal(t, "1.0.0", lockedPkgs["hello"])
	assert.Equal(t, "0.0.1", lockedPkgs["another-pkg"])
}
