package helper

import (
	"os"
	"syscall"

	"golang.org/x/term"
)

func ReadPassword() ([]byte, error) {
	passwd := os.Getenv("CDT_JENKINS_PASSWORD")
	if passwd != "" {
		return []byte(passwd), nil
	}

	return term.ReadPassword(int(syscall.Stdin))
}
