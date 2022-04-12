package helper

import (
	"bytes"
	"io/ioutil"

	"github.com/spf13/cobra"
)

func TestCommand(cmd *cobra.Command, args ...string) (string, error) {
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs(args)
	err := cmd.Execute()

	if err != nil {
		return "", err
	}

	out, err := ioutil.ReadAll(buf)
	if err != nil {
		return "", err
	}

	return string(out), nil
}
