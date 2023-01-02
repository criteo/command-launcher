package repository

import "github.com/criteo/command-launcher/internal/command"

type defaultRepoIndexEntry struct {
	PkgName     string                    `json:"pkgName"`
	PkgVersion  string                    `json:"version"`
	PkgCommands []*command.DefaultCommand `json:"cmds"`
}

func (pkg *defaultRepoIndexEntry) Name() string {
	return pkg.PkgName
}

func (pkg *defaultRepoIndexEntry) Version() string {
	return pkg.PkgVersion
}

func (pkg *defaultRepoIndexEntry) Commands() []command.Command {
	cmds := []command.Command{}
	for _, c := range pkg.PkgCommands {
		cmds = append(cmds, c)
	}
	return cmds
}
