package helper

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCallExternalWithOutput(t *testing.T) {
	cwd, _ := os.Getwd()
	code, output, err := CallExternalWithOutput([]string{}, cwd, "echo", "hello world!")

	assert.Equal(t, 0, code)
	assert.Equal(t, "hello world!\n", output)
	assert.Nil(t, err)
}

func TestCallExternalWithNonZeroExitCode(t *testing.T) {
	cwd, _ := os.Getwd()
	code, _, err := CallExternalWithOutput([]string{}, cwd, "ls", "folder-not-exists")

	assert.NotEqual(t, 0, code)
	assert.NotNil(t, err)
}

func TestCallExternalWithWrongWorkingDirectory(t *testing.T) {
	code, err := CallExternalStdOut([]string{}, "folder-not-exists", "ls")

	assert.NotEqual(t, 0, code)
	assert.NotNil(t, err)
}

func TestCallExternalNoStdOut(t *testing.T) {
	cwd, _ := os.Getwd()
	code, err := CallExternalNoStdOut([]string{}, cwd, "echo", "hello world!")
	// TODO: check no output is in stdout
	assert.Equal(t, 0, code)
	assert.Nil(t, err)

	code, err = CallExternalNoStdOut([]string{}, cwd, "ls", "folder-not-exists")
	// TODO: check no error is show in stderr and stdout
	assert.NotEqual(t, 0, code)
	assert.NotNil(t, err)
}
