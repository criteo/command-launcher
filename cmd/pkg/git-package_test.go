package pkg

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func createGitRepo(t *testing.T) string {
	repoDir := filepath.Join(t.TempDir(), "folder-package")
	err := os.Mkdir(repoDir, 0777)
	assert.Nil(t, err)

	ctx := exec.Command("git", "init")
	ctx.Dir = repoDir
	err = ctx.Run()
	assert.Nil(t, err)

	err = copyFile("assets/folder-package/manifest.mf", filepath.Join(repoDir, "manifest.mf"))
	assert.Nil(t, err)

	ctx = exec.Command("git", "add", "manifest.mf")
	ctx.Dir = repoDir
	err = ctx.Run()
	assert.Nil(t, err)

	ctx = exec.Command("git", "commit", "-m", "initial import")
	ctx.Dir = repoDir
	err = ctx.Run()
	assert.Nil(t, err)

	return repoDir
}

func TestGitRepo_Create_WrongRepo(t *testing.T) {
	p, err := CreateGitRepoPackage("assets/folder-package")
	assert.Nil(t, p)
	assert.NotNil(t, err)
}

func TestGitRepo_Create(t *testing.T) {
	repo := createGitRepo(t)
	p, err := CreateGitRepoPackage(repo)
	assert.NotNil(t, p)
	assert.Nil(t, err)

	assert.Equal(t, 1, len(p.Commands()))
	assert.Equal(t, "fake_test", p.Name())
}

func TestGitRepo_InstallTo(t *testing.T) {
	repo := createGitRepo(t)
	p, err := CreateGitRepoPackage(repo)
	assert.NotNil(t, p)
	assert.Nil(t, err)

	targetDir := t.TempDir()
	mf, err := p.InstallTo(targetDir)
	assert.NotNil(t, mf)
	assert.Nil(t, err)

	_, err = os.Stat(filepath.Join(targetDir, "fake_test", "manifest.mf"))
	assert.Nil(t, err)
}
