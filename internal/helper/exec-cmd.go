package helper

import (
	"fmt"
	"os"
	"os/exec"

	log "github.com/sirupsen/logrus"
)

// Call external command without output to standard out/err
func CallExternalNoStdOut(additionalEnv []string, cwd string, executable string, args ...string) (int, error) {
	if HasDebugFlag(SHOW_CMD_EXEC_STDOUT) {
		return CallExternalStdOut(additionalEnv, cwd, executable, args...)
	}

	code, output, err := CallExternalWithOutput(additionalEnv, cwd, executable, args...)
	log.Debugf("Output command %s", output)
	return code, err
}

// Call external command and output to standard out/err
func CallExternalStdOut(additionalEnv []string, cwd string, executable string, args ...string) (int, error) {
	code, _, err := callExternal(additionalEnv, cwd, false, false, executable, args...)
	return code, err
}

// Call external command and return the output string
func CallExternalWithOutput(additionalEnv []string, cwd string, executable string, args ...string) (int, string, error) {
	return callExternal(additionalEnv, cwd, false, true, executable, args...)
}

func callExternal(additionalEnv []string, cwd string, mute bool, withOutput bool, executable string, args ...string) (int, string, error) {
	if _, err := os.Stat(cwd); os.IsNotExist(err) {
		return 1, "", fmt.Errorf("can't find working directory %s", cwd)
	}

	handle := exec.Command(executable, args...)
	handle.Dir = cwd

	// add additional environments
	env := append(os.Environ(), additionalEnv...)
	handle.Env = env

	if !mute {
		if withOutput {
			output, err := handle.CombinedOutput()
			if err != nil {
				if exitError, ok := err.(*exec.ExitError); ok {
					outtext := ""
					if output != nil {
						outtext = string(output)
					}
					return exitError.ExitCode(), outtext, err
				}
				return 1, "Error when launching the command", err
			}
			return 0, string(output), nil
		} else {
			handle.Stdout = os.Stdout
			handle.Stderr = os.Stderr
			handle.Stdin = os.Stdin
		}
	}

	if err := handle.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return exitError.ExitCode(), "", err
		}

		return -1, "", err
	}

	return 0, "", nil
}
