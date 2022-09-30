package command

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func getDefaultCommand() DefaultCommand {
	return DefaultCommand{
		CmdName:      "test",
		CmdCategory:  "",
		CmdType:      "executable",
		CmdGroup:     "",
		CmdArgsUsage: "[-O opt1] arg1 arg2",
		CmdExamples: []ExampleEntry{
			{
				Scenario: "scenario 1",
				Command:  "test -O opt1 arg1",
			},
		},
		CmdShortDescription:   "test command",
		CmdLongDescription:    "test command - long description",
		CmdExecutable:         "ls",
		CmdArguments:          []string{"-l", "-a"},
		CmdDocFile:            "",
		CmdDocLink:            "",
		CmdValidArgs:          nil,
		CmdValidArgsCmd:       nil,
		CmdRequiredFlags:      nil,
		CmdFlagValuesCmd:      nil,
		CmdCheckFlags:         true,
		CmdRequestedResources: nil,
		PkgDir:                "/tmp/test/root",
	}
}

func TestRequestResources(t *testing.T) {
	cmd := getDefaultCommand()
	assert.NotNil(t, cmd.RequestedResources())
	assert.Equal(t, 0, len(cmd.RequestedResources()))
}

func TestCommandValidArgs(t *testing.T) {
	cmd := getDefaultCommand()
	cmd.CmdValidArgs = []string{"args1", "args2"}
	cmd.CmdValidArgsCmd = []string{}
	cmd.CmdRequiredFlags = []string{"moab", "moab-id"}
	cmd.CmdFlagValuesCmd = []string{}

	validArgs := cmd.ValidArgs()

	assert.Equal(t, 2, len(validArgs))
	assert.Equal(t, "args1", validArgs[0])
	assert.Equal(t, "args2", validArgs[1])

	validArgsCmd := cmd.ValidArgsCmd()
	assert.Equal(t, 0, len(validArgsCmd))

	flagValuesCmd := cmd.FlagValuesCmd()
	assert.Equal(t, 0, len(flagValuesCmd))
}

func TestCommandValidArgsCmd(t *testing.T) {
	cmd := getDefaultCommand()
	cmd.CmdValidArgs = []string{"args1", "args2"}
	cmd.CmdValidArgsCmd = []string{"#CACHE#/test", "arg1", "arg2"}
	cmd.CmdRequiredFlags = []string{"moab", "moab-id"}
	cmd.CmdFlagValuesCmd = []string{"#CACHE#/test", "arg1", "arg2"}

	validArgsCmd := cmd.ValidArgsCmd()
	assert.Equal(t, 3, len(validArgsCmd))
	assert.Equal(t, "#CACHE#/test", validArgsCmd[0])
	assert.Equal(t, "arg1", validArgsCmd[1])
	assert.Equal(t, "arg2", validArgsCmd[2])

	flagValuesCmd := cmd.FlagValuesCmd()
	assert.Equal(t, 3, len(flagValuesCmd))
	assert.Equal(t, "#CACHE#/test", flagValuesCmd[0])
	assert.Equal(t, "arg1", flagValuesCmd[1])
	assert.Equal(t, "arg2", flagValuesCmd[2])

	flags := cmd.RequiredFlags()
	assert.Equal(t, 2, len(flags))
	assert.Equal(t, "moab", flags[0])
	assert.Equal(t, "moab-id", flags[1])
}

func TestNullFields(t *testing.T) {
	cmd := getDefaultCommand()

	assert.NotNil(t, cmd.ValidArgs())
	assert.NotNil(t, cmd.RequiredFlags())
	validArgsCmd := cmd.ValidArgsCmd()
	assert.NotNil(t, validArgsCmd)
	assert.Equal(t, 0, len(validArgsCmd))
}

func TestNewDefaultCommandFromCopy(t *testing.T) {
	cmd := getDefaultCommand()

	newCmd := NewDefaultCommandFromCopy(&cmd, "/new-pkg-dir")
	assert.NotNil(t, newCmd.CmdValidArgs)
	assert.Equal(t, 0, len(newCmd.CmdValidArgs))
	assert.NotNil(t, newCmd.CmdValidArgsCmd)
	assert.Equal(t, 0, len(newCmd.CmdValidArgsCmd))
	assert.NotNil(t, newCmd.CmdRequiredFlags)
	assert.Equal(t, 0, len(newCmd.CmdRequiredFlags))

	assert.Equal(t, 2, len(newCmd.Arguments()))
	assert.Equal(t, "-l", newCmd.Arguments()[0])
	assert.Equal(t, "-a", newCmd.Arguments()[1])
	assert.Equal(t, "/new-pkg-dir", newCmd.PkgDir)

	assert.Equal(t, "[-O opt1] arg1 arg2", newCmd.ArgsUsage())
	assert.Equal(t, 1, len(newCmd.Examples()))
	assert.Equal(t, "scenario 1", newCmd.Examples()[0].Scenario)
	assert.Equal(t, "test -O opt1 arg1", newCmd.Examples()[0].Command)
}

func TestLegacyVariableInterpolation(t *testing.T) {
	cmd := getDefaultCommand()
	cmd.CmdExecutable = "#CACHE#/#OS#/test#EXT#"

	if runtime.GOOS == "windows" {
		assert.Equal(t, fmt.Sprintf("/tmp/test/root/%s/test.exe", runtime.GOOS), cmd.interpolateCmd())
	} else {
		assert.Equal(t, fmt.Sprintf("/tmp/test/root/%s/test", runtime.GOOS), cmd.interpolateCmd())
	}
}

func TestVariableRender(t *testing.T) {
	cmd := getDefaultCommand()
	cmd.CmdExecutable = "{{.Root}}/{{.Os}}/test"

	assert.Equal(t, fmt.Sprintf("/tmp/test/root/%s/test", runtime.GOOS), cmd.interpolateCmd())
}

func TestConditionalVariableRender(t *testing.T) {
	cmd := getDefaultCommand()
	cmd.CmdExecutable = "{{.Root}}/{{.Os}}/test{{if eq .Os \"windows\"}}.bat{{else}}.sh{{end}}"

	if runtime.GOOS == "windows" {
		cmd.PkgDir = "\\tmp\\test\\root"
		assert.Equal(t, fmt.Sprintf("/tmp/test/root/%s/test.bat", runtime.GOOS), cmd.interpolateCmd())
	} else {
		cmd.PkgDir = "/tmp/test/root"
		assert.Equal(t, fmt.Sprintf("/tmp/test/root/%s/test.sh", runtime.GOOS), cmd.interpolateCmd())
	}
}

func TestConditionalBinaryRender(t *testing.T) {
	cmd := getDefaultCommand()
	cmd.CmdExecutable = "{{.Binary}}"

	if runtime.GOOS == "windows" {
		assert.Equal(t, "test.exe", cmd.interpolateCmd())
	} else {
		assert.Equal(t, "test", cmd.interpolateCmd())
	}
}

func TestConditionalScriptRender(t *testing.T) {
	cmd := getDefaultCommand()
	cmd.CmdExecutable = "{{.Script}}"

	if runtime.GOOS == "windows" {
		assert.Equal(t, "test.bat", cmd.interpolateCmd())
	} else {
		assert.Equal(t, "test", cmd.interpolateCmd())
	}
}

func TestConditionalExtensionRender(t *testing.T) {
	cmd := getDefaultCommand()
	cmd.CmdExecutable = "test{{.Extension}}"

	if runtime.GOOS == "windows" {
		assert.Equal(t, "test.exe", cmd.interpolateCmd())
	} else {
		assert.Equal(t, "test", cmd.interpolateCmd())
	}
}

func TestConditionalScriptExtensionRender(t *testing.T) {
	cmd := getDefaultCommand()
	cmd.CmdExecutable = "test{{.ScriptExtension}}"

	if runtime.GOOS == "windows" {
		assert.Equal(t, "test.bat", cmd.interpolateCmd())
	} else {
		assert.Equal(t, "test", cmd.interpolateCmd())
	}
}

func TestMixedRender(t *testing.T) {
	cmd := getDefaultCommand()
	cmd.CmdExecutable = "#CACHE#/#OS#/test{{if eq .Os \"windows\"}}.bat{{else}}.sh{{end}}"

	if runtime.GOOS == "windows" {
		assert.Equal(t, fmt.Sprintf("/tmp/test/root/%s/test.bat", runtime.GOOS), cmd.interpolateCmd())
	} else {
		assert.Equal(t, fmt.Sprintf("/tmp/test/root/%s/test.sh", runtime.GOOS), cmd.interpolateCmd())
	}
}

func TestVariableRenderError(t *testing.T) {
	cmd := getDefaultCommand()
	cmd.CmdExecutable = "{{.Root}}/{{.Os}}/test{{.NonExistKey}}"

	assert.Equal(t, "{{.Root}}/{{.Os}}/test{{.NonExistKey}}", cmd.interpolateCmd())
}

func TestInterpolate(t *testing.T) {
	cmd := getDefaultCommand()
	cmd.CmdExecutable = "#CACHE#/#OS#/#ARCH#/test#EXT#"
	cmd.CmdArguments = []string{"-l", "-a", "#SCRIPT#"}

	assert.Equal(t, ".bat", cmd.doInterpolate("windows", "x64", "#SCRIPT_EXT#"))
	assert.Equal(t, "", cmd.doInterpolate("linux", "x64", "#SCRIPT_EXT#"))
	assert.Equal(t, "test.bat", cmd.doInterpolate("windows", "x64", "#SCRIPT#"))
	assert.Equal(t, "test", cmd.doInterpolate("linux", "x64", "#SCRIPT#"))
	assert.Equal(t, "/tmp/test/root/windows/x64/test.exe", cmd.doInterpolate("windows", "x64", "#CACHE#/#OS#/#ARCH#/test#EXT#"))
}
