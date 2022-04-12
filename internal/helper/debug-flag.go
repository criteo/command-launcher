package helper

import (
	"os"
	"strings"
)

const (
	FORCE_SELF_UPDATE     = "force_self_update"
	NO_MERGE_STATUS_CHECK = "no_merge_status_check"
	SHOW_CMD_EXEC_STDOUT  = "show_cmd_exec_stdout"
	USE_FILE_VAULT        = "use_file_vault"
)

type DebugFlags struct {
	ForceSelfUpdate    bool // Force the self update of the CDT
	NoMergeStatusCheck bool // do not check merge status when querying merged changes in gerrit
	ShowCmdExecStdout  bool // always show cmd exec stdout to console
	UseFileVault       bool // use file vault instead of system vault
}

// load all debug flags into DebugFlags struct
func LoadDebugFlags() DebugFlags {
	flagsString := os.Getenv("CDT_DEBUG_FLAGS")
	flags := strings.Split(flagsString, ",")
	debugFlags := DebugFlags{}
	for _, flag := range flags {
		switch flag {
		case NO_MERGE_STATUS_CHECK:
			debugFlags.NoMergeStatusCheck = true
		case SHOW_CMD_EXEC_STDOUT:
			debugFlags.ShowCmdExecStdout = true
		case FORCE_SELF_UPDATE:
			debugFlags.ForceSelfUpdate = true
		case USE_FILE_VAULT:
			debugFlags.UseFileVault = true
		}
	}
	return debugFlags
}

// check if a debug flag exists
func HasDebugFlag(name string) bool {
	flagsString := os.Getenv("CDT_DEBUG_FLAGS")
	if flagsString == "" {
		return false
	}
	flags := strings.Split(flagsString, ",")
	for _, flag := range flags {
		if flag == name {
			return true
		}
	}
	return false
}
