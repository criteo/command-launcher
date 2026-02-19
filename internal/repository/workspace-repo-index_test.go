package repository

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func createTestPkgDir(t *testing.T, parentDir string, pkgName string, cmdName string, cmdType string) string {
	t.Helper()
	pkgDir := filepath.Join(parentDir, pkgName)
	err := os.MkdirAll(pkgDir, 0755)
	assert.Nil(t, err)

	manifest := []byte(`{
  "pkgName": "` + pkgName + `",
  "version": "1.0.0",
  "cmds": [
    {
      "name": "` + cmdName + `",
      "type": "` + cmdType + `",
      "group": "",
      "short": "test command",
      "executable": "echo"
    }
  ]
}`)
	err = os.WriteFile(filepath.Join(pkgDir, "manifest.mf"), manifest, 0644)
	assert.Nil(t, err)
	return pkgDir
}

func TestWorkspaceRepoIndex_Load(t *testing.T) {
	tmpDir := t.TempDir()

	pkg1Dir := createTestPkgDir(t, tmpDir, "pkg1", "cmd1", "executable")
	pkg2Dir := createTestPkgDir(t, filepath.Join(tmpDir, "tools"), "pkg2", "cmd2", "executable")

	idx, err := NewWorkspaceRepoIndex("workspace:test", []string{pkg1Dir, pkg2Dir})
	assert.Nil(t, err)

	err = idx.Load("")
	assert.Nil(t, err)

	pkgs := idx.AllPackages()
	assert.Len(t, pkgs, 2)

	cmds := idx.ExecutableCommands()
	assert.Len(t, cmds, 2)

	// Verify we can find both packages
	p1, err := idx.Package("pkg1")
	assert.Nil(t, err)
	assert.NotNil(t, p1)

	p2, err := idx.Package("pkg2")
	assert.Nil(t, err)
	assert.NotNil(t, p2)
}

func TestWorkspaceRepoIndex_ReadOnly(t *testing.T) {
	idx, err := NewWorkspaceRepoIndex("workspace:test", []string{})
	assert.Nil(t, err)

	err = idx.Add(nil, "", "")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "read-only")

	err = idx.Remove("foo", "")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "read-only")

	err = idx.Update(nil, "", "")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "read-only")

	err = idx.PausePackageUpdate("foo")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "read-only")
}

func TestWorkspaceRepoIndex_ScatteredPaths(t *testing.T) {
	// Create packages in completely separate temp directories
	dir1 := t.TempDir()
	dir2 := t.TempDir()

	pkg1Dir := createTestPkgDir(t, dir1, "scattered-a", "cmd-a", "executable")
	pkg2Dir := createTestPkgDir(t, dir2, "scattered-b", "cmd-b", "executable")

	idx, err := NewWorkspaceRepoIndex("workspace:scattered", []string{pkg1Dir, pkg2Dir})
	assert.Nil(t, err)

	err = idx.Load("")
	assert.Nil(t, err)

	pkgs := idx.AllPackages()
	assert.Len(t, pkgs, 2)

	cmds := idx.ExecutableCommands()
	assert.Len(t, cmds, 2)

	// Verify PackageDir is set correctly for each command
	for _, cmd := range cmds {
		if cmd.Name() == "cmd-a" {
			assert.Equal(t, pkg1Dir, cmd.PackageDir())
		} else if cmd.Name() == "cmd-b" {
			assert.Equal(t, pkg2Dir, cmd.PackageDir())
		}
	}
}

func TestWorkspaceRepoIndex_InvalidManifest(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a directory with an invalid manifest
	badDir := filepath.Join(tmpDir, "bad-pkg")
	err := os.MkdirAll(badDir, 0755)
	assert.Nil(t, err)
	err = os.WriteFile(filepath.Join(badDir, "manifest.mf"), []byte("not valid json"), 0644)
	assert.Nil(t, err)

	// Create a valid package too
	goodDir := createTestPkgDir(t, tmpDir, "good-pkg", "good-cmd", "executable")

	idx, err := NewWorkspaceRepoIndex("workspace:test", []string{badDir, goodDir})
	assert.Nil(t, err)

	err = idx.Load("")
	assert.Nil(t, err)

	// Should have loaded only the good package
	pkgs := idx.AllPackages()
	assert.Len(t, pkgs, 1)
	assert.Equal(t, "good-pkg", pkgs[0].Name())
}

func TestWorkspaceRepoIndex_IsPackageUpdatePaused(t *testing.T) {
	idx, err := NewWorkspaceRepoIndex("workspace:test", []string{})
	assert.Nil(t, err)

	paused, err := idx.IsPackageUpdatePaused("any")
	assert.Nil(t, err)
	assert.False(t, paused)
}
