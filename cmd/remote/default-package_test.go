package remote

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadManifest(t *testing.T) {
	file, _ := os.Open("assets/fake.mf")
	mf, err := ReadManifest(file)
	assert.Nil(t, err, "cannot read manifest file")

	assert.Equal(t, "fake_test", mf.Name(), "wrong package name")
	assert.Equal(t, "1.0.0", mf.Version(), "wrong package version")

	cmds := mf.Commands()
	assert.NotNil(t, cmds)
	assert.Equal(t, 1, len(cmds))

	assert.Equal(t, "fake_test", cmds[0].Name())
	assert.Equal(t, "Fake manifest", cmds[0].ShortDescription())
	assert.Equal(t, "Fake manifest long description", cmds[0].LongDescription())
	assert.Equal(t, "fake", cmds[0].Executable())
	assert.Equal(t, 2, len(cmds[0].Arguments()))
}

func TestReadManifestInYaml(t *testing.T) {
	file, _ := os.Open("assets/fake-yaml.mf")
	mf, err := ReadManifest(file)
	assert.Nil(t, err, "cannot read manifest file")

	assert.Equal(t, "fake_test", mf.Name(), "wrong package name")
	assert.Equal(t, "1.0.0", mf.Version(), "wrong package version")

	cmds := mf.Commands()
	assert.NotNil(t, cmds)
	assert.Equal(t, 1, len(cmds))

	assert.Equal(t, "fake_test", cmds[0].Name())
	assert.Equal(t, "Fake manifest", cmds[0].ShortDescription())
	assert.Equal(t, "Fake manifest long description\n\nYou can have multiple line descriptions\n", cmds[0].LongDescription())
	assert.Equal(t, "fake", cmds[0].Executable())
	assert.Equal(t, 2, len(cmds[0].Arguments()))
}

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

	target, err := ioutil.TempDir("", "cdt-package-test-*")
	assert.Nil(t, err)

	mf, err := pkg.InstallTo(target)
	assert.Nil(t, err)

	assert.Equal(t, "fake", mf.Name())
	assert.Equal(t, "1.0.0", mf.Version())
	assert.Equal(t, 2, len(mf.Commands()))
}
