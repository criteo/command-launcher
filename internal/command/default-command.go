package command

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/criteo/command-launcher/internal/helper"
	log "github.com/sirupsen/logrus"
)

const (
	CACHE_DIR_PATTERN = "#CACHE#"
	OS_PATTERN        = "#OS#"
	ARCH_PATTERN      = "#ARCH#"
	BINARY_PATTERN    = "#BINARY#"
	EXT_PATTERN       = "#EXT#"
)

/*
DefaultCommand implements the command.Command interface


There are two types of cdt command:
1. group command
2. executable command

A group command doesn't do any thing but contain other executable commands. An executable
command must be under a group command, the default one is the cdt root (group = "")

for example, command: cdt hotfix create

hotfix is a group command, and create is a command under the "hotfix" group command

Another example: cdt ls, here ls is an executable command under the root "" group command

Note: nested group command is not supported! It is not a good practice to have to much level
of nested commands like: cdt workspace create moab.

The group field of group command is ignored.

An additional "category" field is reserved in case we have too much first level commands,
we can use it to category them in the cdt help output.
*/
type DefaultCommand struct {
	CmdName             string   `json:"name"`
	CmdCategory         string   `json:"category"`
	CmdType             string   `json:"type"`
	CmdGroup            string   `json:"group"`
	CmdShortDescription string   `json:"short"`
	CmdLongDescription  string   `json:"long"`
	CmdExecutable       string   `json:"executable"`
	CmdArguments        []string `json:"args"`
	CmdDocFile          string   `json:"docFile"`
	CmdDocLink          string   `json:"docLink"`
	CmdValidArgs        []string `json:"validArgs"`     // the valid argument options
	CmdValidArgsCmd     []string `json:"validArgsCmd"`  // the command to call to get the args for autocompletion
	CmdRequiredFlags    []string `json:"requiredFlags"` // the required flags
	CmdFlagValuesCmd    []string `json:"flagValuesCmd"` // the command to call flag values for autocompletion

	PkgDir string `json:"pkgDir"`
}

func (cmd *DefaultCommand) Execute(envVars []string, args ...string) (int, error) {
	arguments := append(cmd.CmdArguments, args...)
	cmd.interpolateArgs(&arguments)
	command := cmd.interpolateCmd()

	log.Debug("Command line: ", command, " ", arguments)

	ctx := exec.Command(command, arguments...)
	// inject additional environments
	env := append(os.Environ(), envVars...)
	ctx.Env = env

	ctx.Stdout = os.Stdout
	ctx.Stderr = os.Stderr
	ctx.Stdin = os.Stdin

	if err := ctx.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return exitError.ExitCode(), err
		}
	}
	return 0, nil
}

func (cmd *DefaultCommand) ExecuteValidArgsCmd(envVars []string, args ...string) (int, string, error) {
	return cmd.executeArrayCmd(envVars, cmd.CmdValidArgsCmd, args...)
}

func (cmd *DefaultCommand) ExecuteFlagValuesCmd(envVars []string, args ...string) (int, string, error) {
	return cmd.executeArrayCmd(envVars, cmd.CmdFlagValuesCmd, args...)
}

func (cmd *DefaultCommand) executeArrayCmd(envVars []string, cmdArray []string, args ...string) (int, string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return 1, "", err
	}
	validCmd := ""
	validArgs := []string{}
	if cmdArray != nil {
		argsLen := len(cmdArray)
		if argsLen >= 2 {
			validCmd = cmdArray[0]
			validArgs = cmdArray[1:argsLen]
		} else if argsLen >= 1 {
			validCmd = cmdArray[0]
		}
	}
	if validCmd == "" {
		return 0, "", nil
	}
	// Should we interpolate the argumments too???
	//cmd.interpolateArgs(&validArgs)
	//cmd.interpolateArgs(&args)
	return helper.CallExternalWithOutput(envVars, wd, cmd.interpolate(validCmd), append(validArgs, args...)...)
}

func (cmd *DefaultCommand) Name() string {
	return cmd.CmdName
}

func (cmd *DefaultCommand) Type() string {
	if cmd.CmdType != "group" && cmd.CmdType != "executable" {
		// for invalid cmd type, set it to group to make it do nothing
		return "group"
	}
	return cmd.CmdType
}

func (cmd *DefaultCommand) Category() string {
	return cmd.CmdCategory
}

func (cmd *DefaultCommand) Group() string {
	return cmd.CmdGroup
}

func (cmd *DefaultCommand) LongDescription() string {
	return cmd.CmdLongDescription
}

func (cmd *DefaultCommand) ShortDescription() string {
	return cmd.CmdShortDescription
}

func (cmd *DefaultCommand) Executable() string {
	return cmd.CmdExecutable
}

func (cmd *DefaultCommand) Arguments() []string {
	return cmd.CmdArguments
}

func (cmd *DefaultCommand) DocFile() string {
	return cmd.interpolate(cmd.CmdDocFile)
}

func (cmd *DefaultCommand) DocLink() string {
	return cmd.CmdDocLink
}

func (cmd *DefaultCommand) ValidArgs() []string {
	if cmd.CmdValidArgs != nil && len(cmd.CmdValidArgs) > 0 {
		return cmd.CmdValidArgs
	}
	return []string{}
}

func (cmd *DefaultCommand) ValidArgsCmd() []string {
	if cmd.CmdValidArgsCmd != nil && len(cmd.CmdValidArgsCmd) > 0 {
		return cmd.CmdValidArgsCmd
	}
	return []string{}
}

func (cmd *DefaultCommand) RequiredFlags() []string {
	if cmd.CmdRequiredFlags != nil && len(cmd.CmdRequiredFlags) > 0 {
		return cmd.CmdRequiredFlags
	}
	return []string{}
}

func (cmd *DefaultCommand) FlagValuesCmd() []string {
	if cmd.CmdFlagValuesCmd != nil && len(cmd.CmdFlagValuesCmd) > 0 {
		return cmd.CmdFlagValuesCmd
	}
	return []string{}
}

func (cmd *DefaultCommand) Clone() *DefaultCommand {
	return &DefaultCommand{
		CmdName:             cmd.CmdName,
		CmdCategory:         cmd.CmdCategory,
		CmdType:             cmd.CmdType,
		CmdGroup:            cmd.CmdGroup,
		CmdShortDescription: cmd.CmdShortDescription,
		CmdLongDescription:  cmd.CmdLongDescription,
		CmdExecutable:       cmd.CmdExecutable,
		CmdArguments:        cmd.copyArray(cmd.CmdArguments),
		CmdDocFile:          cmd.CmdDocFile,
		CmdDocLink:          cmd.CmdDocLink,
		CmdValidArgs:        cmd.copyArray(cmd.CmdValidArgs),
		CmdValidArgsCmd:     cmd.copyArray(cmd.CmdValidArgsCmd),
		CmdRequiredFlags:    cmd.copyArray(cmd.CmdRequiredFlags),
		CmdFlagValuesCmd:    cmd.copyArray(cmd.CmdFlagValuesCmd),
		PkgDir:              cmd.PkgDir,
	}
}

func (cmd *DefaultCommand) copyArray(src []string) []string {
	if len(src) == 0 {
		return []string{}
	}
	return append([]string{}, src...)
}

func (cmd *DefaultCommand) interpolateArgs(values *[]string) {
	for i := range *values {
		(*values)[i] = strings.ReplaceAll((*values)[i], CACHE_DIR_PATTERN, cmd.PkgDir)
		(*values)[i] = strings.ReplaceAll((*values)[i], OS_PATTERN, runtime.GOOS)
		(*values)[i] = strings.ReplaceAll((*values)[i], ARCH_PATTERN, runtime.GOARCH)
		(*values)[i] = strings.ReplaceAll((*values)[i], BINARY_PATTERN, cmd.binary())
		(*values)[i] = strings.ReplaceAll((*values)[i], EXT_PATTERN, cmd.extension())
	}
}

func (cmd *DefaultCommand) interpolateCmd() string {
	return cmd.interpolate(cmd.CmdExecutable)
}

func (cmd *DefaultCommand) binary() string {
	if runtime.GOOS == "windows" {
		return fmt.Sprintf("%s.exe", cmd.CmdName)
	}
	return cmd.CmdName
}

func (cmd *DefaultCommand) extension() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}

func (cmd *DefaultCommand) interpolate(text string) string {
	output := strings.ReplaceAll(text, CACHE_DIR_PATTERN, cmd.PkgDir)
	output = strings.ReplaceAll(output, OS_PATTERN, runtime.GOOS)
	output = strings.ReplaceAll(output, ARCH_PATTERN, runtime.GOARCH)
	output = strings.ReplaceAll(output, BINARY_PATTERN, cmd.binary())
	output = strings.ReplaceAll(output, EXT_PATTERN, cmd.extension())
	return output
}
