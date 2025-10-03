package pkg

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/criteo/command-launcher/internal/config"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestFolder_Create_EmptyFolder(t *testing.T) {
	p, err := CreateFolderPackage("assets/empty-folder")
	assert.Nil(t, p)
	assert.NotNil(t, err)
}

func TestFolder_Create_WrongManifest(t *testing.T) {
	p, err := CreateFolderPackage("assets/wrong-manifest")
	assert.Nil(t, p)
	assert.NotNil(t, err)
}

func TestFolder_Create_Package(t *testing.T) {
	p, err := CreateFolderPackage("assets/folder-package")
	assert.NotNil(t, p)
	assert.Nil(t, err)

	assert.Equal(t, 1, len(p.Commands()))
	assert.Equal(t, "fake_test", p.Name())
}

func TestFolder_InstallTo(t *testing.T) {
	p, err := CreateFolderPackage("assets/folder-package")
	assert.NotNil(t, p)
	assert.Nil(t, err)

	targetDir := t.TempDir()
	mf, err := p.InstallTo(targetDir)
	assert.NotNil(t, mf)
	assert.Nil(t, err)

	_, err = os.Stat(filepath.Join(targetDir, "fake_test", "manifest.mf"))
	assert.Nil(t, err)
}

func TestFolder_InstallToWithSetupError(t *testing.T) {
	p, err := CreateFolderPackage("assets/fake-wrong-setup")
	assert.NotNil(t, p)
	assert.Nil(t, err)

	targetDir := t.TempDir()
	assert.Nil(t, err)

	var previousValue = viper.GetBool(config.ENABLE_PACKAGE_SETUP_HOOK_KEY)
	viper.Set(config.ENABLE_PACKAGE_SETUP_HOOK_KEY, true)
	mf, err := p.InstallTo(targetDir)
	assert.NotNil(t, err)
	assert.Nil(t, mf)
	viper.Set(config.ENABLE_PACKAGE_SETUP_HOOK_KEY, previousValue)
}
