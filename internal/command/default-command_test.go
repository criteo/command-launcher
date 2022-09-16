package command

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func getDefaultCommand() DefaultCommand {
	return DefaultCommand{
		CmdName:             "test",
		CmdCategory:         "",
		CmdType:             "executable",
		CmdGroup:            "",
		CmdShortDescription: "test command",
		CmdLongDescription:  "test command - long description",
		CmdExecutable:       "ls",
		CmdArguments:        []string{"-l", "-a"},
		CmdDocFile:          "",
		CmdDocLink:          "",
		CmdValidArgs:        nil,
		CmdValidArgsCmd:     nil,
		CmdRequiredFlags:    nil,
		CmdOptionalFlags:    nil,
		CmdExclusiveFlags:   nil,
		CmdTogetherFlags:    nil,
		CmdFlagValuesCmd:    nil,
		PkgDir:              "/tmp/test/root",
	}
}

func TestCommandFlags(t *testing.T) {
	cmd := getDefaultCommand()
	cmd.CmdValidArgs = []string{"args1", "args2"}
	cmd.CmdValidArgsCmd = []string{}
	cmd.CmdRequiredFlags = []string{"moab", "moab-id", "build-num"}
	cmd.CmdOptionalFlags = []string{"opt-1", "opt-2"}
	cmd.CmdExclusiveFlags = [][]string{
		{"moab-id", "build-num"},
	}
	cmd.CmdTogetherFlags = [][]string{
		{"opt-1", "opt-2"},
	}
	cmd.CmdFlagValuesCmd = []string{}

	exclusiveFlags := cmd.ExclusiveFlags()
	assert.Equal(t, 1, len(exclusiveFlags))
	assert.Equal(t, 2, len(exclusiveFlags[0]))
	assert.Equal(t, "moab-id", exclusiveFlags[0][0])
	assert.Equal(t, "build-num", exclusiveFlags[0][1])

	togetherFlags := cmd.TogetherFlags()
	assert.Equal(t, 1, len(togetherFlags))
	assert.Equal(t, 2, len(togetherFlags[0]))
	assert.Equal(t, "opt-1", togetherFlags[0][0])
	assert.Equal(t, "opt-2", togetherFlags[0][1])
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

func TestCloneCommand(t *testing.T) {
	cmd := getDefaultCommand()
	cmd.CmdExecutable = "#CACHE#/#OS#/#ARCH#/test#EXT#"
	cmd.CmdRequiredFlags = []string{"moab", "moab-id", "build-num"}
	cmd.CmdOptionalFlags = []string{"opt-1", "opt-2"}
	cmd.CmdExclusiveFlags = [][]string{
		{"moab-id", "build-num"},
	}
	cmd.CmdTogetherFlags = [][]string{
		{"opt-1", "opt-2"},
	}

	newCmd := cmd.Clone()
	assert.NotNil(t, newCmd.CmdValidArgs)
	assert.Equal(t, 0, len(newCmd.CmdValidArgs))
	assert.NotNil(t, newCmd.CmdValidArgsCmd)
	assert.Equal(t, 0, len(newCmd.CmdValidArgsCmd))
	assert.NotNil(t, newCmd.CmdRequiredFlags)
	assert.Equal(t, 3, len(newCmd.CmdRequiredFlags))

	assert.Equal(t, 2, len(newCmd.Arguments()))
	assert.Equal(t, "-l", newCmd.Arguments()[0])
	assert.Equal(t, "-a", newCmd.Arguments()[1])

	exclusiveFlags := newCmd.ExclusiveFlags()
	assert.Equal(t, 1, len(exclusiveFlags))
	assert.Equal(t, 2, len(exclusiveFlags[0]))
	assert.Equal(t, "moab-id", exclusiveFlags[0][0])
	assert.Equal(t, "build-num", exclusiveFlags[0][1])

	togetherFlags := newCmd.TogetherFlags()
	assert.Equal(t, 1, len(togetherFlags))
	assert.Equal(t, 2, len(togetherFlags[0]))
	assert.Equal(t, "opt-1", togetherFlags[0][0])
	assert.Equal(t, "opt-2", togetherFlags[0][1])

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
