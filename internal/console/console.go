package console

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"github.com/fatih/color"
	"github.com/mattn/go-isatty"
	ps "github.com/mitchellh/go-ps"
	log "github.com/sirupsen/logrus"
)

var (
	isAnsiSequenceSupported = checkAnsiSequenceSupported()
	isTerminal              = fileIsTerminal
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
	// Decide from the environment only; whether the target stream is a
	// terminal is checked per writer, see ansiSupported.
	_, noColor := os.LookupEnv("NO_COLOR")
	maybeSupported := !noColor && os.Getenv("TERM") != "dumb"
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

func fileIsTerminal(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	return isatty.IsTerminal(f.Fd()) || isatty.IsCygwinTerminal(f.Fd())
}

// ansiSupported reports whether ANSI colors and cursor moves can be
// written to w: the environment allows them and w is a terminal.
func ansiSupported(w io.Writer) bool {
	return isAnsiSequenceSupported && isTerminal(w)
}

// Fprogress writes a progress line to w that replaces the previous line
// on a terminal, and a plain line otherwise.
func Fprogress(w io.Writer, format string, a ...interface{}) {
	if ansiSupported(w) {
		fmt.Fprintf(w, "\033[1A"+format+"\033[K\n", a...)
		return
	}
	fmt.Fprintf(w, format+"\n", a...)
}

// ShowCursor makes the cursor of w visible again when it is a terminal.
func ShowCursor(w io.Writer) {
	if ansiSupported(w) {
		fmt.Fprint(w, "\033[?25h")
	}
}

// ClearPreviousLine erases the previous line of w when it is a terminal.
func ClearPreviousLine(w io.Writer) {
	if ansiSupported(w) {
		fmt.Fprint(w, "\033[1A\033[K")
	}
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
	Fhighlight(os.Stdout, format, a...)
}

// Fhighlight writes highlighted informative text to w.
func Fhighlight(w io.Writer, format string, a ...interface{}) {
	printMessage(w, color.New(color.FgBlue), format, a...)
}

// interactive question asked as part of a command's own output,
// for example: ask for consent before running a command
func Prompt(format string, a ...interface{}) {
	Fprompt(os.Stdout, format, a...)
}

// Fprompt writes an interactive question to w.
func Fprompt(w io.Writer, format string, a ...interface{}) {
	printMessage(w, color.New(color.FgMagenta), format, a...)
}

// Diagnostics below write to stderr so command results remain safe to capture.

// usually used as a reminder, or additional information that is
// not part of the command result.
// for example: ask for user to update
func Reminder(format string, a ...interface{}) {
	Fprompt(os.Stderr, format, a...)
}

// usually used as warning message
func Warn(format string, a ...interface{}) {
	printMessage(os.Stderr, color.New(color.FgHiYellow, color.BgBlack), format, a...)
}

// usually used as error message
func Error(format string, a ...interface{}) {
	printMessage(os.Stderr, color.New(color.FgRed), format, a...)
}

// usually used as success message
func Success(format string, a ...interface{}) {
	printMessage(os.Stderr, color.New(color.FgGreen), format, a...)
}

func printMessage(w io.Writer, c *color.Color, format string, a ...interface{}) {
	if !ansiSupported(w) {
		fmt.Fprintf(w, format, a...)
		return
	}
	// color.NoColor reflects stdout only; the decision is made for w.
	c.EnableColor()
	format = addNewline(format)
	message := format
	if len(a) > 0 {
		message = fmt.Sprintf(format, a...)
	}
	// Legacy Windows consoles need ANSI sequences translated to console calls.
	if w == os.Stdout {
		w = color.Output
	} else if w == os.Stderr {
		w = color.Error
	}
	fmt.Fprint(w, c.Sprint(message))
}

func addNewline(format string) string {
	if !strings.HasSuffix(format, "\n") {
		return format + "\n"
	}
	return format
}
