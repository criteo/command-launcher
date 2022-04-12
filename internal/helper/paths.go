package helper

import (
	"path"
	"regexp"
	"runtime"
	"strings"
)

// Check if the path is an absolute path
//
// The standard go function does not support Windows :(
func IsAbsolutePath(pathname string) bool {
	// Path package does not support correctly windows
	abs := path.IsAbs(strings.ReplaceAll(pathname, "\\", "/"))
	if !abs && len(pathname) > 2 && runtime.GOOS == "windows" {
		abs, _ = regexp.Match("[[:alpha:]]", []byte{pathname[0]})
		abs = abs && pathname[1] == ':'
	}

	return abs
}
