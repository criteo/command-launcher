package main

import root "github.com/criteo/command-launcher/cmd"

// Initialized by the linker option (-X main.version=xxxx), this is the build number
// to change the semantic version, see version.go
var version string = "dev"
var buildNum string = "local"
var appName string = "cdt"
var appLongName string = "Criteo Dev Toolkit"

func main() {
	root.InitCommands(appName, appLongName, version, buildNum)
	root.Execute()
}
