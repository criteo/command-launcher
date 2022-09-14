package cmd

import (
	"fmt"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func Test_ParseFlagDefinition(t *testing.T) {
	name, short, desc := parseFlagDefinition("test\t An test flag definition without short form")
	assert.Equal(t, "test", name)
	assert.Equal(t, "", short)
	assert.Equal(t, "An test flag definition without short form", desc)

	name, short, desc = parseFlagDefinition("test\t t \tAn test flag definition with short form")
	assert.Equal(t, "test", name)
	assert.Equal(t, "t", short)
	assert.Equal(t, "An test flag definition with short form", desc)

	name, short, desc = parseFlagDefinition("test")
	assert.Equal(t, "test", name)
	assert.Equal(t, "", short)
	assert.Equal(t, "", desc)
}

func Test_ParseCmdArgToEnv(t *testing.T) {
	cmd := &cobra.Command{
		DisableFlagParsing: true,
		Use:                "cmd",
		Short:              "short",
		Long:               "long",
		Run: func(c *cobra.Command, args []string) {
		},
	}
	cmd.Flags().StringP("flag1", "a", "", "description")
	cmd.Flags().StringP("flag2", "b", "", "description")
	cmd.Flags().BoolP("flag3", "c", false, "description")

	envTable, err := parseCmdArgsToEnv(cmd, []string{"--flag1", "v1", "-b", "v2", "-c", "arg1", "arg2"}, "CDT")
	assert.Nil(t, err)

	assert.True(t, findEnv(envTable, "CDT_FLAG_FLAG1", "v1"))
	assert.True(t, findEnv(envTable, "CDT_FLAG_FLAG2", "v2"))
	assert.True(t, findEnv(envTable, "CDT_FLAG_FLAG3", "true"))

	assert.True(t, findEnv(envTable, "CDT_ARG_1", "arg1"))
	assert.True(t, findEnv(envTable, "CDT_ARG_2", "arg2"))
}

func findEnv(envTable []string, key string, value string) bool {
	for _, v := range envTable {
		if v == fmt.Sprintf("%s=%s", key, value) {
			return true
		}
	}
	return false
}
