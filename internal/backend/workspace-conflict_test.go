package backend

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/criteo/command-launcher/internal/repository"
	"github.com/stretchr/testify/assert"
)

// createTestPackageDir creates a package directory with a manifest.mf file
func createTestPackageDir(t *testing.T, parentDir string, pkgName string, cmds string) string {
	t.Helper()
	pkgDir := filepath.Join(parentDir, pkgName)
	err := os.MkdirAll(pkgDir, 0755)
	assert.Nil(t, err)

	manifest := `{
  "pkgName": "` + pkgName + `",
  "version": "1.0.0",
  "cmds": [` + cmds + `]
}`
	err = os.WriteFile(filepath.Join(pkgDir, "manifest.mf"), []byte(manifest), 0644)
	assert.Nil(t, err)
	return pkgDir
}

func execCmd(name string) string {
	return `{"name": "` + name + `", "type": "executable", "group": "", "short": "test", "executable": "echo"}`
}

func groupCmd(name string) string {
	return `{"name": "` + name + `", "type": "group", "group": "", "short": "test group", "executable": ""}`
}

func execCmdInGroup(name string, group string) string {
	return `{"name": "` + name + `", "type": "executable", "group": "` + group + `", "short": "test", "executable": "echo"}`
}

func makeWorkspaceSource(t *testing.T, dir string, pkgName string, cmds string) *PackageSource {
	t.Helper()
	pkgDir := createTestPackageDir(t, dir, pkgName, cmds)
	repoIndex, err := repository.NewWorkspaceRepoIndex("workspace:"+dir, []string{pkgDir})
	assert.Nil(t, err)
	return &PackageSource{
		Name:            "workspace:" + dir,
		RepoDir:         dir,
		SyncPolicy:      SYNC_POLICY_NEVER,
		IsManaged:       false,
		CustomRepoIndex: repoIndex,
	}
}

func makeDropinSource(t *testing.T, dir string, pkgName string, cmds string) *PackageSource {
	t.Helper()
	createTestPackageDir(t, dir, pkgName, cmds)
	return &PackageSource{
		Name:       "dropin",
		RepoDir:    dir,
		SyncPolicy: SYNC_POLICY_NEVER,
		IsManaged:  false,
	}
}

func makeDefaultSource(t *testing.T, dir string, pkgName string, cmds string) *PackageSource {
	t.Helper()
	createTestPackageDir(t, dir, pkgName, cmds)
	return &PackageSource{
		Name:       "default",
		RepoDir:    dir,
		SyncPolicy: SYNC_POLICY_NEVER,
		IsManaged:  false,
	}
}

func TestWorkspaceOverridesDropin(t *testing.T) {
	homeDir := t.TempDir()
	workspaceDir := t.TempDir()
	dropinDir := t.TempDir()
	defaultDir := t.TempDir()

	workspaceSrc := makeWorkspaceSource(t, workspaceDir, "ws-pkg", execCmd("lint"))
	dropinSrc := makeDropinSource(t, dropinDir, "dropin-pkg", execCmd("lint"))
	defaultSrc := makeDefaultSource(t, defaultDir, "default-pkg", execCmd("deploy"))

	be, err := NewDefaultBackend(homeDir, []*PackageSource{workspaceSrc}, dropinSrc, defaultSrc)
	assert.Nil(t, err)

	// Workspace lint should win the short name
	cmd, err := be.FindCommand("", "lint")
	assert.Nil(t, err)
	assert.Equal(t, "workspace:"+workspaceDir, cmd.RepositoryID())

	// Both commands should be available
	exeCmds := be.ExecutableCommands()
	assert.Len(t, exeCmds, 3) // lint (ws), lint (dropin renamed), deploy
}

func TestWorkspaceOverridesDefault(t *testing.T) {
	homeDir := t.TempDir()
	workspaceDir := t.TempDir()
	dropinDir := t.TempDir()
	defaultDir := t.TempDir()

	// Both define group "build" with executable "run"
	wsCmds := groupCmd("build") + "," + execCmdInGroup("run", "build")
	defCmds := groupCmd("build") + "," + execCmdInGroup("run", "build")

	workspaceSrc := makeWorkspaceSource(t, workspaceDir, "ws-build", wsCmds)
	dropinSrc := makeDropinSource(t, dropinDir, "dropin-pkg", execCmd("other"))
	defaultSrc := makeDefaultSource(t, defaultDir, "default-build", defCmds)

	be, err := NewDefaultBackend(homeDir, []*PackageSource{workspaceSrc}, dropinSrc, defaultSrc)
	assert.Nil(t, err)

	// Workspace build group should win
	cmd, err := be.FindCommand("", "build")
	assert.Nil(t, err)
	assert.Equal(t, "workspace:"+workspaceDir, cmd.RepositoryID())

	// Workspace build run should win
	cmd, err = be.FindCommand("build", "run")
	assert.Nil(t, err)
	assert.Equal(t, "workspace:"+workspaceDir, cmd.RepositoryID())
}

func TestCloserWorkspaceWins(t *testing.T) {
	homeDir := t.TempDir()
	deepDir := t.TempDir()
	shallowDir := t.TempDir()
	dropinDir := t.TempDir()
	defaultDir := t.TempDir()

	// Both workspace levels define "deploy" command
	deepSrc := makeWorkspaceSource(t, deepDir, "deep-pkg", execCmd("deploy"))
	shallowSrc := makeWorkspaceSource(t, shallowDir, "shallow-pkg", execCmd("deploy"))
	dropinSrc := makeDropinSource(t, dropinDir, "dropin-pkg", execCmd("other"))
	defaultSrc := makeDefaultSource(t, defaultDir, "default-pkg", execCmd("yet-another"))

	// Deep source first (closer to cwd)
	be, err := NewDefaultBackend(homeDir, []*PackageSource{deepSrc, shallowSrc}, dropinSrc, defaultSrc)
	assert.Nil(t, err)

	// Deeper workspace should win
	cmd, err := be.FindCommand("", "deploy")
	assert.Nil(t, err)
	assert.Equal(t, "workspace:"+deepDir, cmd.RepositoryID())
}

func TestWorkspaceVsReservedName(t *testing.T) {
	homeDir := t.TempDir()
	workspaceDir := t.TempDir()
	dropinDir := t.TempDir()
	defaultDir := t.TempDir()

	// Workspace defines a group named "config" which is a reserved name
	workspaceSrc := makeWorkspaceSource(t, workspaceDir, "ws-pkg", groupCmd("config"))
	dropinSrc := makeDropinSource(t, dropinDir, "dropin-pkg", execCmd("other"))
	defaultSrc := makeDefaultSource(t, defaultDir, "default-pkg", execCmd("deploy"))

	be, err := NewDefaultBackend(homeDir, []*PackageSource{workspaceSrc}, dropinSrc, defaultSrc)
	assert.Nil(t, err)

	// The workspace config group should have been renamed (not overriding the reserved name)
	groupCmds := be.GroupCommands()
	for _, cmd := range groupCmds {
		if cmd.RepositoryID() == "workspace:"+workspaceDir {
			// It should have been renamed to its full name
			assert.NotEqual(t, "config", cmd.RuntimeName(), "workspace command should not keep reserved name 'config'")
		}
	}
}

func TestNoConflict_DifferentNames(t *testing.T) {
	homeDir := t.TempDir()
	workspaceDir := t.TempDir()
	dropinDir := t.TempDir()
	defaultDir := t.TempDir()

	workspaceSrc := makeWorkspaceSource(t, workspaceDir, "ws-pkg", execCmd("build"))
	dropinSrc := makeDropinSource(t, dropinDir, "dropin-pkg", execCmd("test"))
	defaultSrc := makeDefaultSource(t, defaultDir, "default-pkg", execCmd("deploy"))

	be, err := NewDefaultBackend(homeDir, []*PackageSource{workspaceSrc}, dropinSrc, defaultSrc)
	assert.Nil(t, err)

	// All commands should be accessible by their short names
	cmd, err := be.FindCommand("", "build")
	assert.Nil(t, err)
	assert.Equal(t, "build", cmd.RuntimeName())

	cmd, err = be.FindCommand("", "test")
	assert.Nil(t, err)
	assert.Equal(t, "test", cmd.RuntimeName())

	cmd, err = be.FindCommand("", "deploy")
	assert.Nil(t, err)
	assert.Equal(t, "deploy", cmd.RuntimeName())
}

func TestWorkspaceDisabled_NoWorkspaceSources(t *testing.T) {
	homeDir := t.TempDir()
	dropinDir := t.TempDir()
	defaultDir := t.TempDir()

	dropinSrc := makeDropinSource(t, dropinDir, "dropin-pkg", execCmd("lint"))
	defaultSrc := makeDefaultSource(t, defaultDir, "default-pkg", execCmd("deploy"))

	// No workspace sources (simulating ENABLE_WORKSPACE_PACKAGES=false)
	be, err := NewDefaultBackend(homeDir, []*PackageSource{}, dropinSrc, defaultSrc)
	assert.Nil(t, err)

	// Dropin lint should be available
	cmd, err := be.FindCommand("", "lint")
	assert.Nil(t, err)
	assert.Equal(t, "dropin", cmd.RepositoryID())

	// No workspace sources
	assert.Empty(t, be.WorkspaceSources())
}
