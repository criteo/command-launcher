package repository

import (
	"fmt"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"github.com/criteo/command-launcher/internal/command"
	"github.com/criteo/command-launcher/internal/config"
	"github.com/criteo/command-launcher/internal/pkg"
	"github.com/spf13/viper"
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
	id             string
	packages       map[string]command.PackageManifest
	groupCmds      map[string]command.Command // key is in form of [group]#[cmd name] ex. "#hotfix"
	executableCmds map[string]command.Command // key is in form of [group]#[cmd name] ex. "hotfix#create"
	systemCmds     map[string]command.Command // key is the predefined system command name
}

func newDefaultRegistry(id string) (Registry, error) {
	reg := defaultRegistry{
		id:             id,
		packages:       make(map[string]command.PackageManifest),
		groupCmds:      make(map[string]command.Command),
		executableCmds: make(map[string]command.Command),
		systemCmds:     make(map[string]command.Command),
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

		sysPkgName := viper.GetString(config.SYSTEM_PACKAGE_KEY)
		for _, f := range files {
			if !f.IsDir() && f.Type()&os.ModeSymlink != os.ModeSymlink {
				continue
			}
			if manifestFile, err := os.Open(filepath.Join(repoDir, f.Name(), "manifest.mf")); err == nil {
				defer manifestFile.Close()
				manifest, err := pkg.ReadManifest(manifestFile)
				if err == nil {
					reg.packages[manifest.Name()] = manifest
					for _, cmd := range manifest.Commands() {
						cmd.SetPackageDir(filepath.Join(repoDir, f.Name()))
						cmd.SetNamespace(reg.id, manifest.Name())
						reg.registerCmd(manifest, cmd, sysPkgName != "" && sysPkgName == manifest.Name())
					}
				}
			}
		}
	}

	return err
}

func (reg *defaultRegistry) Add(pkg command.PackageManifest, repoDir string) error {
	reg.packages[pkg.Name()] = pkg
	reg.extractCmds(repoDir)
	return nil
}

func (reg *defaultRegistry) Remove(pkgName string, repoDir string) error {
	delete(reg.packages, pkgName)
	reg.extractCmds(repoDir)
	return nil
}

func (reg *defaultRegistry) Update(pkg command.PackageManifest, repoDir string) error {
	reg.packages[pkg.Name()] = pkg
	reg.extractCmds(repoDir)
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
	if cmd, exist := reg.groupCmds[fmt.Sprintf("%s#%s", group, name)]; exist {
		return cmd, nil
	}

	if cmd, exist := reg.executableCmds[fmt.Sprintf("%s#%s", group, name)]; exist {
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
		cmds = append(cmds, v)
	}
	return cmds
}

func (reg *defaultRegistry) SystemLoginCommand() command.Command {
	if cmd, exist := reg.systemCmds[SYSTEM_LOGIN_COMMAND]; exist {
		return cmd
	}
	return nil
}

func (reg *defaultRegistry) SystemMetricsCommand() command.Command {
	if cmd, exist := reg.systemCmds[SYSTEM_METRICS_COMMAND]; exist {
		return cmd
	}
	return nil
}

func (reg *defaultRegistry) ExecutableCommands() []command.Command {
	cmds := make([]command.Command, 0)
	for _, v := range reg.executableCmds {
		cmds = append(cmds, v)
	}
	return cmds
}

func (reg *defaultRegistry) extractCmds(repoDir string) {
	sysPkgName := viper.GetString(config.SYSTEM_PACKAGE_KEY)
	reg.groupCmds = make(map[string]command.Command)
	reg.executableCmds = make(map[string]command.Command)
	// initiate group cmds and exectuable cmds map
	// the key is in format of [group]#[cmd name]
	for _, pkg := range reg.packages {
		if pkg.Commands() != nil {
			for _, cmd := range pkg.Commands() {
				cmd.SetPackageDir(filepath.Join(repoDir, pkg.Name()))
				cmd.SetNamespace(reg.id, pkg.Name())
				reg.registerCmd(pkg, cmd, sysPkgName != "" && pkg.Name() == sysPkgName)
			}
		}
	}
}

func (reg *defaultRegistry) registerCmd(pkg command.PackageManifest, cmd command.Command, isSystemPkg bool) {
	switch cmd.Type() {
	case "group":
		reg.groupCmds[fmt.Sprintf("#%s", cmd.Name())] = cmd
	case "executable":
		reg.executableCmds[fmt.Sprintf("%s#%s", cmd.Group(), cmd.Name())] = cmd
	case "system":
		if isSystemPkg {
			reg.extractSystemCmds(cmd)
		}
	}
}

func (reg *defaultRegistry) extractSystemCmds(cmd command.Command) {
	switch cmd.Name() {
	case SYSTEM_LOGIN_COMMAND:
		reg.systemCmds[SYSTEM_LOGIN_COMMAND] = cmd
	case SYSTEM_METRICS_COMMAND:
		reg.systemCmds[SYSTEM_METRICS_COMMAND] = cmd
	}
}
