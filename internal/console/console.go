package console

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/fatih/color"
	ps "github.com/mitchellh/go-ps"
	log "github.com/sirupsen/logrus"
)

var (
	colorWarn               = color.New(color.FgHiYellow).Add(color.BgBlack)
	isAnsiSequenceSupported = checkAnsiSequenceSupported()
)

func processName() string {
	if runtime.GOOS == "windows" {
		return "cdt.exe"
	}

	return "cdt"
}

func parentName() (string, error) {
	parent, err := ps.FindProcess(os.Getppid())
	if err != nil {
		return "", err
	}

	if parent.Executable() == processName() {
		parent, err = ps.FindProcess(parent.PPid())
		if err != nil {
			return "", err
		}
	}

	return parent.Executable(), nil
}

func checkAnsiSequenceSupported() bool {
	maybeSupported := !color.NoColor
	// The Powershell console does not support the escape ANSI chars
	if runtime.GOOS == "windows" && maybeSupported {
		parent, err := parentName()
		if err != nil {
			maybeSupported = false
		} else {
			log.Debugf("Parent Process name %s", parent)
			maybeSupported = strings.ToLower(parent) != "powershell.exe"
		}
	}

	log.Debugf("ANSI sequences are supported by the console: %t", maybeSupported)
	return maybeSupported
}

// When the ANSI sequence is supported, the console will display the colors and cursor moves
//
// On Windows, when cdt runs in a powershell console, the ANSI sequences are not supported.
func IsAnsiSequenceSupported() bool {
	return isAnsiSequenceSupported
}

// Color code
//
// Blue: for highlighted informative text
// Magenta: for interactive questions, reminder
// Yellow: for warning
// Red: for errors

// highlight informative text in the output so that
// user can distinguish different steps in the output
// for example: highlight the command used under the hood
// see: hotfix create, hotfix review
func Highlight(format string, a ...interface{}) {
	if isAnsiSequenceSupported {
		color.Blue(format, a...)
	} else {
		fmt.Printf(format, a...)
	}
}

// usually used as a reminder, a question, or additional
// information that need the user interact, but not mandatory.
// for example: ask for user to update
func Reminder(format string, a ...interface{}) {
	if isAnsiSequenceSupported {
		color.Magenta(format, a...)
	} else {
		fmt.Printf(format, a...)
	}
}

// usually used as warning message
func Warn(format string, a ...interface{}) {
	if isAnsiSequenceSupported {
		if !strings.HasSuffix(format, "\n") {
			format += "\n"
		}
		colorWarn.Printf(format, a...)
	} else {
		fmt.Printf(format, a...)
	}
}

// usually used as error message
func Error(format string, a ...interface{}) {
	if isAnsiSequenceSupported {
		color.Red(format, a...)
	} else {
		fmt.Printf(format, a...)
	}
}

// usually used as success message
func Success(format string, a ...interface{}) {
	if isAnsiSequenceSupported {
		color.Green(format, a...)
	} else {
		fmt.Printf(format, a...)
	}
}
