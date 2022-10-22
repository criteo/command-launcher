package repository

import (
	"fmt"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"github.com/criteo/command-launcher/cmd/remote"
	"github.com/criteo/command-launcher/internal/command"
)

/*
Internal data structure to represent the local registry
You can find an example of the file in the path specified by
config "local_command_repository_dirname"

Current implementation of the registry is to store each package as
an entry (into a single json file), and load commands into memory
during startup

Further improvements could store registry as indexes, and laze load
further information when necessary to reduce the startup time
*/
type defaultRegistry struct {
	packages       map[string]command.PackageManifest
	groupCmds      map[string]command.Command // key is in form of [group]_[cmd name] ex. "_hotfix"
	executableCmds map[string]command.Command // key is in form of [group]_[cmd name] ex. "hotfix_create"
}

func newDefaultRegistry() (Registry, error) {
	reg := defaultRegistry{
		packages:       make(map[string]command.PackageManifest),
		groupCmds:      make(map[string]command.Command),
		executableCmds: make(map[string]command.Command),
	}

	return &reg, nil
}

func (reg *defaultRegistry) Load(repoDir string) error {
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
			if dropinPkgManifestFile, err := os.Open(filepath.Join(repoDir, f.Name(), "manifest.mf")); err == nil {
				manifest, err := remote.ReadManifest(dropinPkgManifestFile)
				if err == nil {
					for _, cmd := range manifest.Commands() {
						newCmd := command.NewDefaultCommandFromCopy(cmd, filepath.Join(repoDir, f.Name()))
						if newCmd.CmdType == "group" {
							reg.groupCmds[fmt.Sprintf("_%s", cmd.Name())] = newCmd
						} else {
							reg.executableCmds[fmt.Sprintf("%s_%s", cmd.Group(), cmd.Name())] = newCmd
						}
					}
				}
			}
		}
	}

	return err
}

func (reg *defaultRegistry) Add(pkg command.PackageManifest) error {
	reg.packages[pkg.Name()] = pkg
	reg.extractCmds()
	return nil
}

func (reg *defaultRegistry) Remove(pkgName string) error {
	delete(reg.packages, pkgName)
	reg.extractCmds()
	return nil
}

func (reg *defaultRegistry) Update(pkg command.PackageManifest) error {
	reg.packages[pkg.Name()] = pkg
	reg.extractCmds()
	return nil
}

func (reg *defaultRegistry) AllPackages() []command.PackageManifest {
	pkgs := []command.PackageManifest{}
	for _, p := range reg.packages {
		newPkg := p
		pkgs = append(pkgs, newPkg)
	}
	return pkgs
}

func (reg *defaultRegistry) Package(name string) (command.PackageManifest, error) {
	if pkg, exists := reg.packages[name]; exists {
		return pkg, nil
	}
	return nil, fmt.Errorf("cannot find the package '%s'", name)
}

func (reg *defaultRegistry) Command(group string, name string) (command.Command, error) {
	if cmd, exist := reg.groupCmds[fmt.Sprintf("%s_%s", group, name)]; exist {
		return cmd, nil
	}

	if cmd, exist := reg.executableCmds[fmt.Sprintf("%s_%s", group, name)]; exist {
		return cmd, nil
	}

	return nil, fmt.Errorf("cannot find the command %s %s", group, name)
}

func (reg *defaultRegistry) AllCommands() []command.Command {
	cmds := reg.GroupCommands()
	cmds = append(cmds, reg.ExecutableCommands()...)
	return cmds
}

func (reg *defaultRegistry) GroupCommands() []command.Command {
	cmds := make([]command.Command, 0)

	for _, v := range reg.groupCmds {
		//groupCmd := v
		cmds = append(cmds, v)
	}

	return cmds
}

func (reg *defaultRegistry) ExecutableCommands() []command.Command {
	cmds := make([]command.Command, 0)

	for _, v := range reg.executableCmds {
		//exeCmd := v
		cmds = append(cmds, v)
	}

	return cmds
}

func (reg *defaultRegistry) extractCmds() {
	reg.groupCmds = make(map[string]command.Command)
	reg.executableCmds = make(map[string]command.Command)
	// initiate group cmds and exectuable cmds map
	// the key is in format of [group]_[cmd name]
	for _, pkg := range reg.packages {
		if pkg.Commands() != nil {
			for _, cmd := range pkg.Commands() {
				newCmd := cmd
				if cmd.Type() == "group" {
					reg.groupCmds[fmt.Sprintf("_%s", cmd.Name())] = newCmd
				} else {
					reg.executableCmds[fmt.Sprintf("%s_%s", cmd.Group(), cmd.Name())] = newCmd
				}
			}
		}
	}
}
