package pkg

import (
	"os"
	"testing"

	"github.com/criteo/command-launcher/internal/config"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestCreatePackage(t *testing.T) {
	pkg, err := CreateZipPackage("assets/fake-1.0.0.pkg")
	assert.Nil(t, err, "cannot create package")

	assert.Equal(t, "fake", pkg.Name())
	assert.Equal(t, "1.0.0", pkg.Version())
	assert.Equal(t, 2, len(pkg.Commands()))
}

func TestInstallPackage(t *testing.T) {
	pkg, err := CreateZipPackage("assets/fake-1.0.0.pkg")
	assert.Nil(t, err)

	target, err := os.MkdirTemp("", "cdt-package-test-*")
	assert.Nil(t, err)

	mf, err := pkg.InstallTo(target)
	assert.Nil(t, err)

	assert.Equal(t, "fake", mf.Name())
	assert.Equal(t, "1.0.0", mf.Version())
	assert.Equal(t, 2, len(mf.Commands()))
}

func TestInstallPackageWithSetupError(t *testing.T) {
	pkg, err := CreateZipPackage("assets/fake-wrong-setup-1.0.0.pkg")
	assert.Nil(t, err)

	target, err := os.MkdirTemp("", "cdt-package-test-*")
	assert.Nil(t, err)

	var previousValue = viper.GetBool(config.ENABLE_PACKAGE_SETUP_HOOK_KEY)
	viper.Set(config.ENABLE_PACKAGE_SETUP_HOOK_KEY, true)
	mf, err := pkg.InstallTo(target)
	assert.NotNil(t, err)
	assert.Nil(t, mf)
	viper.Set(config.ENABLE_PACKAGE_SETUP_HOOK_KEY, previousValue)
}

func TestVerifyChecksum(t *testing.T) {
	pkg, err := CreateZipPackage("assets/fake-1.0.0.pkg")
	assert.Nil(t, err)
	verified, err := pkg.VerifyChecksum("353b23600bd2c3a661c6b825b2a27f19ee14938903bac24290ec26a5c9fa5bb4")
	assert.Nil(t, err)
	assert.True(t, verified)
}
