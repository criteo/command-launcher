package dropin

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/criteo/command-launcher/cmd/remote"
	"github.com/criteo/command-launcher/internal/command"
	log "github.com/sirupsen/logrus"
)

type DropinRepository struct {
	groupCmds      map[string]*command.DefaultCommand // key is in form of [group]_[cmd name] ex. "_hotfix"
	executableCmds map[string]*command.DefaultCommand // key is in form of [group]_[cmd name] ex. "hotfix_create"
}

func Load(pathname string) (*DropinRepository, error) {
	registry := DropinRepository{
		groupCmds:      make(map[string]*command.DefaultCommand),
		executableCmds: make(map[string]*command.DefaultCommand),
	}

	_, err := os.Stat(pathname)
	if !os.IsNotExist(err) {
		files, err := ioutil.ReadDir(pathname)

		if err != nil {
			log.Fatal(err)
			return &registry, err
		}

		for _, f := range files {
			if !f.IsDir() && f.Mode()&os.ModeSymlink != os.ModeSymlink {
				continue
			}
			if dropinPkgManifestFile, err := os.Open(filepath.Join(pathname, f.Name(), "manifest.mf")); err == nil {
				manifest, err := remote.ReadManifest(dropinPkgManifestFile)
				if err == nil {
					for _, cmd := range manifest.Commands() {
						newCmd := command.DefaultCommand{
							CmdName:             cmd.Name(),
							CmdCategory:         cmd.Category(),
							CmdType:             cmd.Type(),
							CmdGroup:            cmd.Group(),
							CmdShortDescription: cmd.ShortDescription(),
							CmdLongDescription:  cmd.LongDescription(),
							CmdExecutable:       cmd.Executable(),
							CmdArguments:        cmd.Arguments(),
							CmdDocFile:          cmd.DocFile(),
							CmdDocLink:          cmd.DocLink(),
							CmdValidArgs:        cmd.ValidArgs(),
							CmdValidArgsCmd:     cmd.ValidArgsCmd(),
							CmdRequiredFlags:    cmd.RequiredFlags(),
							CmdFlagValuesCmd:    cmd.FlagValuesCmd(),
							CmdCheckFlags:       cmd.CheckFlags(),
							PkgDir:              filepath.Join(pathname, f.Name()),
						}
						if newCmd.CmdType == "group" {
							registry.groupCmds[fmt.Sprintf("_%s", cmd.Name())] = &newCmd
						} else {
							registry.executableCmds[fmt.Sprintf("%s_%s", cmd.Group(), cmd.Name())] = &newCmd
						}
					}
				}
			}
		}
	}

	return &registry, nil
}

func (reg *DropinRepository) GroupCommands() []command.Command {
	cmds := make([]command.Command, 0)

	for _, v := range reg.groupCmds {
		cmds = append(cmds, v)
	}

	return cmds
}

func (reg *DropinRepository) ExecutableCommands() []command.Command {
	cmds := make([]command.Command, 0)

	for _, v := range reg.executableCmds {
		cmds = append(cmds, v)
	}

	return cmds
}

func (reg *DropinRepository) Command(group string, name string) (command.Command, error) {
	if cmd, exist := reg.groupCmds[fmt.Sprintf("%s_%s", group, name)]; exist {
		return cmd, nil
	}

	if cmd, exist := reg.executableCmds[fmt.Sprintf("%s_%s", group, name)]; exist {
		return cmd, nil
	}

	return nil, fmt.Errorf("cannot find the command %s %s", group, name)
}
