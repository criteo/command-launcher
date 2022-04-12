package helper

import (
	"os"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"
)

func ReadPassword() ([]byte, error) {
	passwd := os.Getenv("CDT_JENKINS_PASSWORD")
	if passwd != "" {
		return []byte(passwd), nil
	}

	return terminal.ReadPassword(int(syscall.Stdin))
}
