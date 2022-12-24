package repository

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/criteo/command-launcher/internal/helper"
	"github.com/criteo/command-launcher/internal/remote"
	"github.com/stretchr/testify/assert"
)

func TestLocalRepository(t *testing.T) {
	localRepoPath := strings.ReplaceAll(filepath.Join(t.TempDir(), "local-repo-test"), "\\", "/")
	err := os.Mkdir(localRepoPath, 0755)
	assert.Nil(t, err)

	regFile, err := os.Create(filepath.Join(localRepoPath, "registry.json"))
	assert.Nil(t, err)
	defer regFile.Close()
	regFile.WriteString(fmt.Sprintf(`{
		"ls": {
			"pkgName": "ls",
			"version": "0.0.2",
			"cmds": [
				{"name": "ls",
				"category": "",
				"type": "executable",
				"group": "",
				"short": "A wrapper of linux command 'ls'",
				"long": "A wrapper of linux command 'ls'",
				"executable": "ls",
				"args": ["-l", "-a"],
				"docFile": "",
				"docLink": "",
				"validArgs": null,
				"validArgsCmd": null,
				"requiredFlags": null,
				"pkgDir": "%s/%s"}
			]
		}
	}`, localRepoPath, "ls-0.0.2"))

	reg, err := newJsonRegistry(filepath.Join(localRepoPath, "registry.json"))
	assert.Nil(t, err)

	localRepo, err := CreateLocalRepository(localRepoPath, reg)
	assert.Nil(t, err)

	ls, err := localRepo.Command("", "ls")
	assert.Nil(t, err)
	assert.Equal(t, "ls", ls.Name())
	assert.Equal(t, "executable", ls.Type())

	cmds := localRepo.InstalledCommands()
	assert.Equal(t, 1, len(cmds))
	assert.Equal(t, "ls", cmds[0].Name())

	pkgs := localRepo.InstalledPackages()
	assert.Equal(t, 1, len(pkgs))
	assert.Equal(t, "ls", pkgs[0].Name())
	assert.Equal(t, "0.0.2", pkgs[0].Version())

	execCmds := localRepo.InstalledExecutableCommands()
	assert.Equal(t, 1, len(execCmds))
	assert.Equal(t, "ls", execCmds[0].Name())

	groupCmds := localRepo.InstalledGroupCommands()
	assert.Equal(t, 0, len(groupCmds))

}

func TestInstallCommand(t *testing.T) {
	// init remote
	basePath := filepath.Join(t.TempDir(), "remote-test")
	err := os.Mkdir(basePath, 0755)
	assert.Nil(t, err)

	indexPath := filepath.Join(basePath, "index.json")
	err = helper.CopyLocalFile("../remote/assets/remote/basic-index.json", indexPath, false)
	assert.Nil(t, err)

	err = helper.CopyLocalFile("../remote/assets/ls-0.0.3.pkg", filepath.Join(basePath, "ls-0.0.3.pkg"), false)
	assert.Nil(t, err)

	err = helper.CopyLocalFile("../remote/assets/ls-0.0.2.pkg", filepath.Join(basePath, "ls-0.0.2.pkg"), false)
	assert.Nil(t, err)

	remoteRepo := remote.CreateRemoteRepository(fmt.Sprintf("file://%s", basePath))
	remoteRepo.Fetch()

	// init local repo
	localRepoPath := filepath.Join(t.TempDir(), "local-repo-test")
	err = os.Mkdir(localRepoPath, 0755)
	assert.Nil(t, err)

	reg, err := newJsonRegistry(filepath.Join(localRepoPath, "registry.json"))
	assert.Nil(t, err)

	localRepo, err := CreateLocalRepository(localRepoPath, reg)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(localRepo.InstalledCommands()))

	// install package
	lsPkg, err := remoteRepo.Package("ls", "0.0.2")
	assert.Nil(t, err)
	err = localRepo.Install(lsPkg)
	assert.Nil(t, err)

	installedCmds := localRepo.InstalledCommands()
	assert.Equal(t, 1, len(installedCmds))
	assert.Equal(t, "ls", installedCmds[0].Name())

	installedPkgs := localRepo.InstalledPackages()
	assert.Equal(t, 1, len(installedPkgs))
	assert.Equal(t, "ls", installedPkgs[0].Name())
	assert.Equal(t, "0.0.2", installedPkgs[0].Version())

	pkgManifest, err := localRepo.Package("ls")
	assert.Nil(t, err)
	assert.Equal(t, "ls", pkgManifest.Name())
	assert.Equal(t, "0.0.2", pkgManifest.Version())

	// upadte it
	lsV3, err := remoteRepo.Package("ls", "0.0.3")
	assert.Nil(t, err)

	err = localRepo.Update(lsV3)
	assert.Nil(t, err)

	pkgManifest, err = localRepo.Package("ls")
	assert.Nil(t, err)
	assert.Equal(t, "ls", pkgManifest.Name())
	assert.Equal(t, "0.0.3", pkgManifest.Version())

	// now uninstall it
	err = localRepo.Uninstall("ls")
	assert.Nil(t, err)
	assert.Equal(t, 0, len(localRepo.InstalledCommands()))
}

func Test_Load(t *testing.T) {
	pathname, err := filepath.Abs("assets/simple_dropins/")
	if err == nil {
		fmt.Println("Absolute:", pathname)
	}

	reg, err := newDefaultRegistry()
	assert.Nil(t, err)

	repo, err := CreateLocalRepository(pathname, reg)
	assert.Nil(t, err)

	assert.Equal(t, 1, len(repo.InstalledGroupCommands()))
	assert.Equal(t, 1, len(repo.InstalledExecutableCommands()))

	assert.Equal(t, "wf", repo.InstalledGroupCommands()[0].Name())
	assert.Equal(t, "debug-cdt-env", repo.InstalledExecutableCommands()[0].Name())
}

func Test_Load_Unexist_Folder(t *testing.T) {
	pathname, err := filepath.Abs("assets/simple_dropins_not_exist/")
	if err == nil {
		fmt.Println("Absolute:", pathname)
	}

	reg, err := newDefaultRegistry()
	assert.Nil(t, err)

	repo, err := CreateLocalRepository(pathname, reg)
	assert.Nil(t, err)

	assert.Equal(t, 0, len(repo.InstalledGroupCommands()))
	assert.Equal(t, 0, len(repo.InstalledExecutableCommands()))
}

func Test_Load_Malformat_Manifest(t *testing.T) {
	pathname, err := filepath.Abs("assets/dropins_wrong_manifest_format/")
	if err == nil {
		fmt.Println("Absolute:", pathname)
	}

	reg, err := newDefaultRegistry()
	assert.Nil(t, err)

	repo, err := CreateLocalRepository(pathname, reg)
	assert.Nil(t, err)

	assert.Equal(t, 0, len(repo.InstalledGroupCommands()))
	assert.Equal(t, 0, len(repo.InstalledExecutableCommands()))

}

func Test_Load_Multiple_Pkgs(t *testing.T) {
	pathname, err := filepath.Abs("assets/dropins_multiple_pkgs/")
	if err == nil {
		fmt.Println("Absolute:", pathname)
	}

	reg, err := newDefaultRegistry()
	assert.Nil(t, err)

	repo, err := CreateLocalRepository(pathname, reg)
	assert.Nil(t, err)

	assert.Equal(t, 2, len(repo.InstalledGroupCommands()))
	assert.Equal(t, 2, len(repo.InstalledExecutableCommands()))
}

func Test_Load_Symlink(t *testing.T) {
	pathname, err := filepath.Abs("assets/symlink_dropins/")
	if err == nil {
		fmt.Println("Absolute:", pathname)
	}

	reg, err := newDefaultRegistry()
	assert.Nil(t, err)

	repo, err := CreateLocalRepository(pathname, reg)
	assert.Nil(t, err)

	assert.Equal(t, 1, len(repo.InstalledGroupCommands()))
	assert.Equal(t, 1, len(repo.InstalledExecutableCommands()))

	assert.Equal(t, "wf", repo.InstalledGroupCommands()[0].Name())
	assert.Equal(t, "debug-cdt-env", repo.InstalledExecutableCommands()[0].Name())
}
