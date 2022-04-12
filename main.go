package main

import root "github.com/criteo/command-launcher/cmd"

// Initialized by the linker option (-X main.version=xxxx), this is the build number
// to change the semantic version, see version.go
var version string = ""
var binaryName string = ""

func init() {
	root.BuildNum = version
}

func main() {
	root.Execute()
}
