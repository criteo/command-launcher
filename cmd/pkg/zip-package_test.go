package pkg

import (
	"os"
	"testing"

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
