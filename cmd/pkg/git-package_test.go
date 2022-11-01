package pkg

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/criteo/command-launcher/internal/helper"
	"github.com/stretchr/testify/assert"
)

func createGitRepo(t *testing.T) string {
	repoDir := filepath.Join(t.TempDir(), "folder-package")
	err := os.Mkdir(repoDir, 0777)
	assert.Nil(t, err)

	commands := [...][]string{
		{"init"},
		{"add", "manifest.mf"},
		{"config", "user.email", "test@example.com"},
		{"config", "user.name", "tester"},
		{"commit", "-m", "initial import"},
	}

	err = helper.CopyLocalFile("assets/folder-package/manifest.mf", filepath.Join(repoDir, "manifest.mf"), false)
	assert.Nil(t, err)

	for _, cmd := range commands {
		ctx := exec.Command("git", cmd...)
		ctx.Dir = repoDir
		ctx.Stdout = os.Stdout
		ctx.Stderr = os.Stderr
		ctx.Stdin = os.Stdin
		err = ctx.Run()
		assert.Nil(t, err)
	}

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
