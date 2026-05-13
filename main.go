package main

import (
	"os"
	"path/filepath"
	"strings"

	root "github.com/criteo/command-launcher/cmd"
)

// Initialized by the linker option (-X main.version=xxxx), this is the build number
// to change the semantic version, see version.go
var version string = "dev"
var buildNum string = "local"
var appName string = "cdt"
var appLongName string = "Criteo Dev Toolkit"

// resolveAppName derives the application name from the real path of the
// running binary. Symlinks are resolved so that symbolic links behave as
// aliases (same binary identity), while copies or hard links produce a
// distinct name and therefore a separate configuration directory.
// Falls back to the compiled-in appName on any error.
func resolveAppName() string {
	exe, err := os.Executable()
	if err != nil {
		return appName
	}
	resolved, err := filepath.EvalSymlinks(exe)
	if err != nil {
		resolved = exe
	}
	name := filepath.Base(resolved)
	name = strings.TrimSuffix(name, filepath.Ext(name))
	if name == "" || name == "." {
		return appName
	}
	return name
}

func main() {
	runtimeAppName := resolveAppName()
	root.InitCommands(runtimeAppName, appLongName, version, buildNum)
	root.Execute()
}
