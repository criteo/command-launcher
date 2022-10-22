package remote

import (
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
