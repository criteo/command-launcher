package main

import (
	"fmt"
	"os"
	"runtime"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println(runtime.GOOS)
		os.Exit(0)
	}

	if args[0] == "extension" {
		if runtime.GOOS == "windows" {
			fmt.Println(".exe")
		} else {
			fmt.Println("")
		}
	}
}
