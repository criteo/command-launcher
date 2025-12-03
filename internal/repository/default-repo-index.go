package repository

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	log "github.com/sirupsen/logrus"

	"github.com/criteo/command-launcher/internal/command"
	"github.com/criteo/command-launcher/internal/config"
	"github.com/criteo/command-launcher/internal/pkg"
	"github.com/criteo/command-launcher/internal/updateConfig"
	"github.com/spf13/viper"
)

/*
Internal data structure to represent the local repository index
You can find an example of the file in the path specified by
config "local_command_repository_dirname"

Current implementation of the repoIndex is to scan all manifest.mf files
in one level subfolders

Further improvements could store commands as indexes, and laze load
further information when necessary to reduce the startup time
*/
const (
	PACKAGE_UPDATE_FILE = ".update"
)

type defaultRepoIndex struct {
	id             string
	packages       map[string]command.PackageManifest
	packageDirs    map[string]string          // key is the package name, value is the package directory
	groupCmds      map[string]command.Command // key is in form of [repo]>[package]>[group]>[cmd name]
	executableCmds map[string]command.Command // key is in form of [repo]>[package]>[group]>[cmd name]
	systemCmds     map[string]command.Command // key is the predefined system command name
}

func newDefaultRepoIndex(id string) (RepoIndex, error) {
	repoIndex := defaultRepoIndex{
		id:             id,
		packages:       make(map[string]command.PackageManifest),
		packageDirs:    make(map[string]string),
		groupCmds:      map[string]command.Command{},
		executableCmds: map[string]command.Command{},
		systemCmds:     make(map[string]command.Command),
	}

	return &repoIndex, nil
}

func (repoIndex *defaultRepoIndex) loadPackages(repoDir string) error {
	_, err := os.Stat(repoDir)
	if !os.IsNotExist(err) {
		files, err := os.ReadDir(repoDir)
		if err != nil {
			log.Errorf("cannot read the repo dir: %v", err)
			return err
		}

		for _, f := range files {
			if !f.IsDir() && f.Type()&os.ModeSymlink != os.ModeSymlink {
				continue
			}
			if manifestFile, err := os.Open(filepath.Join(repoDir, f.Name(), "manifest.mf")); err == nil {
				defer manifestFile.Close()
				manifest, err := pkg.ReadManifest(manifestFile)
				if err == nil {
					repoIndex.packages[manifest.Name()] = manifest
					repoIndex.packageDirs[manifest.Name()] = filepath.Join(repoDir, f.Name())
				}
			}
		}
	}
	return err
}

func (repoIndex *defaultRepoIndex) extractCmds(repoDir string) {
	sysPkgName := viper.GetString(config.SYSTEM_PACKAGE_KEY)
	repoIndex.groupCmds = make(map[string]command.Command)
	repoIndex.executableCmds = make(map[string]command.Command)
	// initiate group cmds and exectuable cmds map
	// the key is in format of [group]#[cmd name]
	for _, pkg := range repoIndex.packages {
		if pkg.Commands() != nil {
			for _, cmd := range pkg.Commands() {
				cmd.SetPackageDir(repoIndex.packageDirs[pkg.Name()])
				cmd.SetNamespace(repoIndex.id, pkg.Name())
				repoIndex.registerCmd(pkg, cmd,
					sysPkgName != "" && pkg.Name() == sysPkgName && repoIndex.id == "default", // always use default repository for system package
				)
			}
		}
	}
}

func (repoIndex *defaultRepoIndex) Load(repoDir string) error {
	err := repoIndex.loadPackages(repoDir)
	if err != nil {
		return err
	}
	repoIndex.extractCmds(repoDir)
	return nil
}

func (repoIndex *defaultRepoIndex) Add(pkg command.PackageManifest, repoDir string, pkgDirName string) error {
	repoIndex.packages[pkg.Name()] = pkg
	repoIndex.packageDirs[pkg.Name()] = filepath.Join(repoDir, pkgDirName)
	repoIndex.extractCmds(repoDir)
	return nil
}

func (repoIndex *defaultRepoIndex) Remove(pkgName string, repoDir string) error {
	delete(repoIndex.packages, pkgName)
	delete(repoIndex.packageDirs, pkgName)
	repoIndex.extractCmds(repoDir)
	return nil
}

func (repoIndex *defaultRepoIndex) Update(pkg command.PackageManifest, repoDir string, pkgDirName string) error {
	repoIndex.packages[pkg.Name()] = pkg
	repoIndex.packageDirs[pkg.Name()] = filepath.Join(repoDir, pkgDirName)
	repoIndex.extractCmds(repoDir)
	return nil
}

func (repoIndex *defaultRepoIndex) AllPackages() []command.PackageManifest {
	pkgs := []command.PackageManifest{}
	for _, p := range repoIndex.packages {
		newPkg := p
		pkgs = append(pkgs, newPkg)
	}
	return pkgs
}

func (repoIndex *defaultRepoIndex) Package(name string) (command.PackageManifest, error) {
	if pkg, exists := repoIndex.packages[name]; exists {
		return pkg, nil
	}
	return nil, fmt.Errorf("cannot find the package '%s'", name)
}

func (repoIndex *defaultRepoIndex) IsPackageUpdatePaused(name string) (bool, error) {
	if _, exists := repoIndex.packageDirs[name]; !exists {
		return false, fmt.Errorf("cannot find the package '%s'", name)
	} else {
		pkgDir := repoIndex.packageDirs[name]
		exists, err := updateConfig.IsUpdateConfigExists(pkgDir)
		if err != nil {
			return false, err
		}
		if !exists {
			return false, nil
		}
		uConfig, err := updateConfig.ReadFromDir(pkgDir)
		if err != nil {
			return false, err
		}
		return uConfig.IsExpired(), nil
	}
}

func (repoIndex *defaultRepoIndex) PausePackageUpdate(name string) error {
	if _, exists := repoIndex.packageDirs[name]; !exists {
		return fmt.Errorf("cannot find the package '%s'", name)
	} else {
		pkgDir := repoIndex.packageDirs[name]
		uConfig := &updateConfig.UpdateConfig{}
		if exists, err := updateConfig.IsUpdateConfigExists(pkgDir); err != nil {
			return err
		} else if exists {
			uConfig, err = updateConfig.ReadFromDir(pkgDir)
			if err != nil {
				return err
			}
		}
		uConfig.UpdateAfterDate(updateConfig.DEFAULT_UPDATE_LOCK_DURATION)
		return uConfig.WriteToDir(pkgDir)
	}
}

func (repoIndex *defaultRepoIndex) Command(pkg string, group string, name string) (command.Command, error) {
	if cmd, exist := repoIndex.groupCmds[command.CmdID(repoIndex.id, pkg, group, name)]; exist {
		return cmd, nil
	}

	if cmd, exist := repoIndex.executableCmds[command.CmdID(repoIndex.id, pkg, group, name)]; exist {
		return cmd, nil
	}

	return nil, fmt.Errorf("cannot find the command %s %s", group, name)
}

func (repoIndex *defaultRepoIndex) AllCommands() []command.Command {
	cmds := repoIndex.GroupCommands()
	cmds = append(cmds, repoIndex.ExecutableCommands()...)
	sort.Slice(cmds, func(i, j int) bool {
		return cmds[i].ID() < cmds[j].ID()
	})
	return cmds
}

func (repoIndex *defaultRepoIndex) GroupCommands() []command.Command {
	cmds := make([]command.Command, 0)
	for _, v := range repoIndex.groupCmds {
		cmds = append(cmds, v)
	}
	sort.Slice(cmds, func(i, j int) bool {
		return cmds[i].ID() < cmds[j].ID()
	})
	return cmds
}

func (repoIndex *defaultRepoIndex) SystemLoginCommand() command.Command {
	if cmd, exist := repoIndex.systemCmds[SYSTEM_LOGIN_COMMAND]; exist {
		return cmd
	}
	return nil
}

func (repoIndex *defaultRepoIndex) SystemMetricsCommand() command.Command {
	if cmd, exist := repoIndex.systemCmds[SYSTEM_METRICS_COMMAND]; exist {
		return cmd
	}
	return nil
}

func (repoIndex *defaultRepoIndex) ExecutableCommands() []command.Command {
	cmds := make([]command.Command, 0)
	for _, v := range repoIndex.executableCmds {
		cmds = append(cmds, v)
	}
	sort.Slice(cmds, func(i, j int) bool {
		return cmds[i].ID() < cmds[j].ID()
	})
	return cmds
}

func (repoIndex *defaultRepoIndex) registerCmd(pkg command.PackageManifest, cmd command.Command, isSystemPkg bool) {
	switch cmd.Type() {
	case "group":
		repoIndex.groupCmds[cmd.ID()] = cmd
	case "executable":
		repoIndex.executableCmds[cmd.ID()] = cmd
	case "system":
		if isSystemPkg {
			repoIndex.extractSystemCmds(cmd)
		}
	}
}

func (repoIndex *defaultRepoIndex) extractSystemCmds(cmd command.Command) {
	switch cmd.Name() {
	case SYSTEM_LOGIN_COMMAND:
		repoIndex.systemCmds[SYSTEM_LOGIN_COMMAND] = cmd
	case SYSTEM_METRICS_COMMAND:
		repoIndex.systemCmds[SYSTEM_METRICS_COMMAND] = cmd
	}
}
