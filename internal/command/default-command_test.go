package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommandValidArgs(t *testing.T) {
	cmd := DefaultCommand{
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
		CmdValidArgs:        []string{"args1", "args2"},
		CmdValidArgsCmd:     []string{},
		CmdRequiredFlags:    []string{"moab", "moab-id"},
		CmdFlagValuesCmd:    []string{},
		PkgDir:              "/tmp/test/root",
	}

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
	cmd := DefaultCommand{
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
		CmdValidArgs:        []string{"args1", "args2"},
		CmdValidArgsCmd:     []string{"#CACHE#/test", "arg1", "arg2"},
		CmdRequiredFlags:    []string{"moab", "moab-id"},
		CmdFlagValuesCmd:    []string{"#CACHE#/test", "arg1", "arg2"},
		PkgDir:              "/tmp/test/root",
	}

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
	cmd := DefaultCommand{
		CmdName:             "test",
		CmdCategory:         "",
		CmdType:             "executable",
		CmdGroup:            "",
		CmdShortDescription: "test command",
		CmdLongDescription:  "test command - long description",
		CmdExecutable:       "#CACHE#/#OS#/#ARCH#/test#EXT#",
		CmdArguments:        []string{"-l", "-a"},
		CmdDocFile:          "",
		CmdDocLink:          "",
		CmdValidArgs:        nil,
		CmdValidArgsCmd:     nil,
		CmdRequiredFlags:    nil,
		CmdFlagValuesCmd:    nil,
		PkgDir:              "/tmp/test/root",
	}

	assert.NotNil(t, cmd.ValidArgs())
	assert.NotNil(t, cmd.RequiredFlags())
	validArgsCmd := cmd.ValidArgsCmd()
	assert.NotNil(t, validArgsCmd)
	assert.Equal(t, 0, len(validArgsCmd))
}

func TestCloneCommand(t *testing.T) {
	cmd := DefaultCommand{
		CmdName:             "test",
		CmdCategory:         "",
		CmdType:             "executable",
		CmdGroup:            "",
		CmdShortDescription: "test command",
		CmdLongDescription:  "test command - long description",
		CmdExecutable:       "#CACHE#/#OS#/#ARCH#/test#EXT#",
		CmdArguments:        []string{"-l", "-a"},
		CmdDocFile:          "",
		CmdDocLink:          "",
		CmdValidArgs:        nil,
		CmdValidArgsCmd:     nil,
		CmdRequiredFlags:    nil,
		CmdFlagValuesCmd:    nil,
		PkgDir:              "/tmp/test/root",
	}

	newCmd := cmd.Clone()
	assert.NotNil(t, newCmd.CmdValidArgs)
	assert.Equal(t, 0, len(newCmd.CmdValidArgs))
	assert.NotNil(t, newCmd.CmdValidArgsCmd)
	assert.Equal(t, 0, len(newCmd.CmdValidArgsCmd))
	assert.NotNil(t, newCmd.CmdRequiredFlags)
	assert.Equal(t, 0, len(newCmd.CmdRequiredFlags))

	assert.Equal(t, 2, len(newCmd.Arguments()))
	assert.Equal(t, "-l", newCmd.Arguments()[0])
	assert.Equal(t, "-a", newCmd.Arguments()[1])
}
