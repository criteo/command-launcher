package repository

import "github.com/criteo/command-launcher/internal/command"

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
