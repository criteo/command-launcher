package dropin

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Dropin_Load(t *testing.T) {
	dropinsPath, err := filepath.Abs("assets/simple_dropins/")
	if err == nil {
		fmt.Println("Absolute:", dropinsPath)
	}
	repo, err := Load(dropinsPath)
	assert.Nil(t, err)

	assert.Equal(t, 1, len(repo.GroupCommands()))
	assert.Equal(t, 1, len(repo.ExecutableCommands()))

	assert.Equal(t, "wf", repo.GroupCommands()[0].Name())
	assert.Equal(t, "debug-cdt-env", repo.ExecutableCommands()[0].Name())
}

func Test_Dropin_Load_Unexist_Folder(t *testing.T) {
	dropinsPath, err := filepath.Abs("assets/simple_dropins_not_exist/")
	if err == nil {
		fmt.Println("Absolute:", dropinsPath)
	}
	repo, err := Load(dropinsPath)
	assert.Nil(t, err)

	assert.Equal(t, 0, len(repo.GroupCommands()))
	assert.Equal(t, 0, len(repo.ExecutableCommands()))
}

func Test_Dropin_Load_Malformat_Manifest(t *testing.T) {
	dropinsPath, err := filepath.Abs("assets/dropins_wrong_manifest_format/")
	if err == nil {
		fmt.Println("Absolute:", dropinsPath)
	}
	repo, err := Load(dropinsPath)
	assert.Nil(t, err)

	assert.Equal(t, 0, len(repo.GroupCommands()))
	assert.Equal(t, 0, len(repo.ExecutableCommands()))

}

func Test_Dropin_Load_Multiple_Pkgs(t *testing.T) {
	dropinsPath, err := filepath.Abs("assets/dropins_multiple_pkgs/")
	if err == nil {
		fmt.Println("Absolute:", dropinsPath)
	}
	repo, err := Load(dropinsPath)
	assert.Nil(t, err)

	assert.Equal(t, 2, len(repo.GroupCommands()))
	assert.Equal(t, 2, len(repo.ExecutableCommands()))
}

func Test_Dropin_Load_Symlink(t *testing.T) {
	dropinsPath, err := filepath.Abs("assets/symlink_dropins/")
	if err == nil {
		fmt.Println("Absolute:", dropinsPath)
	}
	repo, err := Load(dropinsPath)
	assert.Nil(t, err)

	assert.Equal(t, 1, len(repo.GroupCommands()))
	assert.Equal(t, 1, len(repo.ExecutableCommands()))

	assert.Equal(t, "wf", repo.GroupCommands()[0].Name())
	assert.Equal(t, "debug-cdt-env", repo.ExecutableCommands()[0].Name())
}
