package repository

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/criteo/command-launcher/internal/command"
	"github.com/criteo/command-launcher/internal/config"
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
	packages       map[string]defaultRegistryEntry
	groupCmds      map[string]*command.DefaultCommand // key is in form of [group]_[cmd name] ex. "_hotfix"
	executableCmds map[string]*command.DefaultCommand // key is in form of [group]_[cmd name] ex. "hotfix_create"
	systemCmds     map[string]command.Command         // key is the predefined system command name
}

type defaultRegistryEntry struct {
	PkgName     string                    `json:"pkgName"`
	PkgVersion  string                    `json:"version"`
	PkgCommands []*command.DefaultCommand `json:"cmds"`
}

func (pkg *defaultRegistryEntry) Name() string {
	return pkg.PkgName
}

func (pkg *defaultRegistryEntry) Version() string {
	return pkg.PkgVersion
}

func (pkg *defaultRegistryEntry) Commands() []command.Command {
	cmds := []command.Command{}
	for _, c := range pkg.PkgCommands {
		cmds = append(cmds, c)
	}
	return cmds
}

func NewRegistryEntry(pkg command.Package, pkgDir string) defaultRegistryEntry {
	defPkg := defaultRegistryEntry{
		PkgName:     pkg.Name(),
		PkgVersion:  pkg.Version(),
		PkgCommands: []*command.DefaultCommand{},
	}

	for _, cmd := range pkg.Commands() {
		newCmd := command.NewDefaultCommandFromCopy(cmd, pkgDir)
		defPkg.PkgCommands = append(defPkg.PkgCommands, newCmd)
	}

	return defPkg
}

func LoadRegistry(pathname string) (*defaultRegistry, error) {
	registry := defaultRegistry{
		packages:       make(map[string]defaultRegistryEntry),
		groupCmds:      make(map[string]*command.DefaultCommand),
		executableCmds: make(map[string]*command.DefaultCommand),
		systemCmds:     make(map[string]command.Command),
	}

	_, err := os.Stat(pathname)
	if !os.IsNotExist(err) {
		payload, err := ioutil.ReadFile(pathname)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(payload, &registry.packages)
		if err != nil {
			return nil, err
		}
	}

	registry.extractCmds()

	return &registry, nil
}

func (reg *defaultRegistry) Store(pathname string) error {
	payload, err := json.Marshal(reg.packages)
	if err != nil {
		return fmt.Errorf("cannot encode in json: %v", err)
	}

	err = ioutil.WriteFile(pathname, payload, 0755)
	if err != nil {
		return fmt.Errorf("cannot write registry file: %v", err)
	}

	return nil
}

func (reg *defaultRegistry) Add(pkg defaultRegistryEntry) error {
	reg.packages[pkg.PkgName] = pkg
	reg.extractCmds()
	return nil
}

func (reg *defaultRegistry) Remove(pkgName string) error {
	delete(reg.packages, pkgName)
	reg.extractCmds()
	return nil
}

func (reg *defaultRegistry) Update(pkg defaultRegistryEntry) error {
	reg.packages[pkg.PkgName] = pkg
	reg.extractCmds()
	return nil
}

func (reg *defaultRegistry) AllPackages() []command.PackageManifest {
	pkgs := []command.PackageManifest{}
	for _, p := range reg.packages {
		newPkg := p
		pkgs = append(pkgs, &newPkg)
	}
	return pkgs
}

func (reg *defaultRegistry) Package(name string) (command.PackageManifest, error) {
	if pkg, exists := reg.packages[name]; exists {
		return &pkg, nil
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
		//exeCmd := v
		cmds = append(cmds, v)
	}

	return cmds
}

func (reg *defaultRegistry) extractCmds() {
	sysPkgName := viper.GetString(config.SYSTEM_PACKAGE_KEY)

	reg.groupCmds = make(map[string]*command.DefaultCommand)
	reg.executableCmds = make(map[string]*command.DefaultCommand)
	// initiate group cmds and exectuable cmds map
	// the key is in format of [group]_[cmd name]
	for _, pkg := range reg.packages {
		if pkg.PkgCommands != nil {
			for _, cmd := range pkg.PkgCommands {
				newCmd := cmd
				switch cmd.CmdType {
				case "group":
					reg.groupCmds[fmt.Sprintf("_%s", cmd.Name())] = newCmd
				case "executable":
					reg.executableCmds[fmt.Sprintf("%s_%s", cmd.Group(), cmd.Name())] = newCmd
				case "system":
					if sysPkgName != "" && pkg.Name() == sysPkgName {
						reg.extractSystemCmds(newCmd)
					}
				}
			}
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
