package main

import root "github.com/criteo/command-launcher/cmd"

// Initialized by the linker option (-X main.version=xxxx), this is the build number
// to change the semantic version, see version.go
var version string = "dev"
var appName string = "cdt"
var appLongName string = "Criteo Dev Toolkit"

func init() {
	root.BuildVersion = version
	root.AppName = appName
	root.LongAppName = appLongName
}

func main() {
	root.Execute()
}
