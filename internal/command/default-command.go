package command

import (
	"fmt"
	"html/template"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/criteo/command-launcher/internal/helper"
	log "github.com/sirupsen/logrus"
)

const (
	CACHE_DIR_PATTERN  = "#CACHE#"
	OS_PATTERN         = "#OS#"
	ARCH_PATTERN       = "#ARCH#"
	BINARY_PATTERN     = "#BINARY#"
	SCRIPT_PATTERN     = "#SCRIPT#"
	EXT_PATTERN        = "#EXT#"
	SCRIPT_EXT_PATTERN = "#SCRIPT_EXT#"
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
	CmdName             string   `json:"name" yaml:"name"`
	CmdCategory         string   `json:"category" yaml:"category"`
	CmdType             string   `json:"type" yaml:"type"`
	CmdGroup            string   `json:"group" yaml:"group"`
	CmdShortDescription string   `json:"short" yaml:"short"`
	CmdLongDescription  string   `json:"long" yaml:"long"`
	CmdExecutable       string   `json:"executable" yaml:"executable"`
	CmdArguments        []string `json:"args" yaml:"args"`
	CmdDocFile          string   `json:"docFile" yaml:"docFile"`
	CmdDocLink          string   `json:"docLink" yaml:"docLink"`
	CmdValidArgs        []string `json:"validArgs" yaml:"validArgs"`         // the valid argument options
	CmdValidArgsCmd     []string `json:"validArgsCmd" yaml:"validArgsCmd"`   // the command to call to get the args for autocompletion
	CmdRequiredFlags    []string `json:"requiredFlags" yaml:"requiredFlags"` // the required flags
	CmdFlagValuesCmd    []string `json:"flagValuesCmd" yaml:"flagValuesCmd"` // the command to call flag values for autocompletion
	CmdCheckFlags       bool     `json:"checkFlags" yaml:"checkFlags"`       // whether parse the flags and check them before execution

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
		log.Debug("Command execution err: ", err)
		if exitError, ok := err.(*exec.ExitError); ok {
			log.Debug("Exit code: ", exitError.ExitCode())
			return exitError.ExitCode(), err
		}
	}
	log.Debug("Command executed successfully")
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
	cmd.interpolateArgs(&validArgs)
	// Should we interpolate the argumments too???
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

func (cmd *DefaultCommand) CheckFlags() bool {
	return cmd.CmdCheckFlags
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
		CmdCheckFlags:       cmd.CmdCheckFlags,
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
		(*values)[i] = cmd.interpolate((*values)[i])
	}
}

func (cmd *DefaultCommand) interpolateCmd() string {
	return cmd.interpolate(cmd.CmdExecutable)
}

func (cmd *DefaultCommand) binary(os string) string {
	if os == "windows" {
		return fmt.Sprintf("%s.exe", cmd.CmdName)
	}
	return cmd.CmdName
}

func (cmd *DefaultCommand) extension(os string) string {
	if os == "windows" {
		return ".exe"
	}
	return ""
}

func (cmd *DefaultCommand) script(os string) string {
	return fmt.Sprintf("%s%s", cmd.CmdName, cmd.script_ext(os))
}

func (cmd *DefaultCommand) script_ext(os string) string {
	if os == "windows" {
		return ".bat"
	}
	return ""
}

func (cmd *DefaultCommand) interpolate(text string) string {
	return cmd.doInterpolate(runtime.GOOS, runtime.GOARCH, text)
}

func (cmd *DefaultCommand) doInterpolate(os string, arch string, text string) string {
	output := strings.ReplaceAll(text, CACHE_DIR_PATTERN, filepath.ToSlash(cmd.PkgDir))
	output = strings.ReplaceAll(output, OS_PATTERN, os)
	output = strings.ReplaceAll(output, ARCH_PATTERN, arch)
	output = strings.ReplaceAll(output, BINARY_PATTERN, cmd.binary(os))
	output = strings.ReplaceAll(output, EXT_PATTERN, cmd.extension(os))
	output = strings.ReplaceAll(output, SCRIPT_PATTERN, cmd.script(os))
	output = strings.ReplaceAll(output, SCRIPT_EXT_PATTERN, cmd.script_ext(os))
	output = cmd.render(output)
	return output
}

// Support golang built-in text/template engine
type TemplateContext struct {
	Os              string
	Arch            string
	Cache           string
	Root            string
	Binary          string
	Script          string
	Extension       string
	ScriptExtension string
}

func (cmd *DefaultCommand) render(text string) string {
	ctx := TemplateContext{
		Os:              runtime.GOOS,
		Arch:            runtime.GOARCH,
		Cache:           filepath.ToSlash(cmd.PkgDir),
		Root:            filepath.ToSlash(cmd.PkgDir),
		Binary:          cmd.binary(runtime.GOOS),
		Script:          cmd.script(runtime.GOOS),
		Extension:       cmd.extension(runtime.GOOS),
		ScriptExtension: cmd.script_ext(runtime.GOOS),
	}

	t, err := template.New("command-template").Parse(text)
	if err != nil {
		return text
	}

	builder := strings.Builder{}
	err = t.Execute(&builder, ctx)
	if err != nil {
		return text
	}

	return builder.String()
}
