package backend

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func createTestManifest(t *testing.T, dir string, pkgName string) {
	t.Helper()
	pkgDir := filepath.Join(dir, pkgName)
	err := os.MkdirAll(pkgDir, 0755)
	assert.Nil(t, err)
	manifest := []byte(`{
  "pkgName": "` + pkgName + `",
  "version": "1.0.0",
  "cmds": [
    {
      "name": "` + pkgName + `-cmd",
      "type": "executable",
      "group": "",
      "short": "test command",
      "executable": "echo"
    }
  ]
}`)
	err = os.WriteFile(filepath.Join(pkgDir, "manifest.mf"), manifest, 0644)
	assert.Nil(t, err)
}

func TestDiscoverWorkspaceSources_NoneFound(t *testing.T) {
	tmpDir := t.TempDir()
	sources := DiscoverWorkspaceSources(tmpDir)
	assert.Empty(t, sources)
}

func TestDiscoverWorkspaceSources_SingleFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create project structure
	projectDir := filepath.Join(tmpDir, "project")
	err := os.MkdirAll(projectDir, 0755)
	assert.Nil(t, err)

	// Create a package
	createTestManifest(t, projectDir, "my-tool")

	// Create .cdt-packages
	err = os.WriteFile(filepath.Join(projectDir, WorkspacePackagesFileName), []byte("my-tool\n"), 0644)
	assert.Nil(t, err)

	// Start from a subdirectory
	subDir := filepath.Join(projectDir, "src", "deep")
	err = os.MkdirAll(subDir, 0755)
	assert.Nil(t, err)

	sources := DiscoverWorkspaceSources(subDir)
	assert.Len(t, sources, 1)
	assert.Equal(t, WorkspaceSourcePrefix+projectDir, sources[0].Name)
	assert.Equal(t, projectDir, sources[0].RepoDir)
	assert.False(t, sources[0].IsManaged)
	assert.Equal(t, SYNC_POLICY_NEVER, sources[0].SyncPolicy)
}

func TestDiscoverWorkspaceSources_MultipleFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create outer workspace with a package
	workspaceDir := filepath.Join(tmpDir, "workspace")
	err := os.MkdirAll(workspaceDir, 0755)
	assert.Nil(t, err)
	createTestManifest(t, workspaceDir, "shared-tool")
	err = os.WriteFile(filepath.Join(workspaceDir, WorkspacePackagesFileName), []byte("shared-tool\n"), 0644)
	assert.Nil(t, err)

	// Create inner project with a package
	projectDir := filepath.Join(workspaceDir, "my-project")
	err = os.MkdirAll(projectDir, 0755)
	assert.Nil(t, err)
	createTestManifest(t, projectDir, "project-tool")
	err = os.WriteFile(filepath.Join(projectDir, WorkspacePackagesFileName), []byte("project-tool\n"), 0644)
	assert.Nil(t, err)

	// Start from project subdirectory
	subDir := filepath.Join(projectDir, "src")
	err = os.MkdirAll(subDir, 0755)
	assert.Nil(t, err)

	sources := DiscoverWorkspaceSources(subDir)
	assert.Len(t, sources, 2)
	// Deepest first
	assert.Equal(t, WorkspaceSourcePrefix+projectDir, sources[0].Name)
	assert.Equal(t, WorkspaceSourcePrefix+workspaceDir, sources[1].Name)
}

func TestParseWorkspaceFile_CommentsAndBlanks(t *testing.T) {
	tmpDir := t.TempDir()

	createTestManifest(t, tmpDir, "valid-pkg")

	content := `# This is a comment

valid-pkg

# Another comment
`
	dotFile := filepath.Join(tmpDir, WorkspacePackagesFileName)
	err := os.WriteFile(dotFile, []byte(content), 0644)
	assert.Nil(t, err)

	paths, err := ParseWorkspaceFile(dotFile)
	assert.Nil(t, err)
	assert.Len(t, paths, 1)
	assert.Equal(t, filepath.Join(tmpDir, "valid-pkg"), paths[0])
}

func TestParseWorkspaceFile_RelativePaths(t *testing.T) {
	tmpDir := t.TempDir()

	// Create package in a nested directory
	toolsDir := filepath.Join(tmpDir, "tools", "my-tool")
	err := os.MkdirAll(toolsDir, 0755)
	assert.Nil(t, err)
	createTestManifest(t, filepath.Join(tmpDir, "tools"), "my-tool")

	content := "tools/my-tool\n"
	dotFile := filepath.Join(tmpDir, WorkspacePackagesFileName)
	err = os.WriteFile(dotFile, []byte(content), 0644)
	assert.Nil(t, err)

	paths, err := ParseWorkspaceFile(dotFile)
	assert.Nil(t, err)
	assert.Len(t, paths, 1)
	assert.Equal(t, filepath.Join(tmpDir, "tools", "my-tool"), paths[0])
}

func TestParseWorkspaceFile_InvalidPath(t *testing.T) {
	tmpDir := t.TempDir()

	content := "nonexistent-package\n"
	dotFile := filepath.Join(tmpDir, WorkspacePackagesFileName)
	err := os.WriteFile(dotFile, []byte(content), 0644)
	assert.Nil(t, err)

	paths, err := ParseWorkspaceFile(dotFile)
	assert.Nil(t, err)
	assert.Empty(t, paths)
}

func TestParseWorkspaceFile_RejectParentTraversal(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a package outside the workspace
	outsideDir := filepath.Join(tmpDir, "outside")
	createTestManifest(t, outsideDir, "evil-pkg")

	// Create workspace inside
	workspaceDir := filepath.Join(tmpDir, "workspace")
	err := os.MkdirAll(workspaceDir, 0755)
	assert.Nil(t, err)

	content := `../outside/evil-pkg
tools/../../outside/evil-pkg
`
	dotFile := filepath.Join(workspaceDir, WorkspacePackagesFileName)
	err = os.WriteFile(dotFile, []byte(content), 0644)
	assert.Nil(t, err)

	paths, err := ParseWorkspaceFile(dotFile)
	assert.Nil(t, err)
	assert.Empty(t, paths, "paths with .. should be rejected")
}

func TestParseWorkspaceFile_AllowDotPaths(t *testing.T) {
	tmpDir := t.TempDir()

	// Create package with dot prefix path
	createTestManifest(t, filepath.Join(tmpDir, "tools"), "my-tool")

	// Create package in a hidden directory
	hiddenDir := filepath.Join(tmpDir, ".hidden")
	createTestManifest(t, hiddenDir, "hidden-tool")

	content := `./tools/my-tool
.hidden/hidden-tool
`
	dotFile := filepath.Join(tmpDir, WorkspacePackagesFileName)
	err := os.WriteFile(dotFile, []byte(content), 0644)
	assert.Nil(t, err)

	paths, err := ParseWorkspaceFile(dotFile)
	assert.Nil(t, err)
	assert.Len(t, paths, 2)
}

func TestContainsParentTraversal(t *testing.T) {
	assert.True(t, containsParentTraversal("../foo"))
	assert.True(t, containsParentTraversal("foo/../../bar"))
	assert.True(t, containsParentTraversal(".."))
	assert.False(t, containsParentTraversal("./foo"))
	assert.False(t, containsParentTraversal("foo/bar"))
	assert.False(t, containsParentTraversal(".hidden/pkg"))
	assert.False(t, containsParentTraversal("tools/my-tool"))
}
