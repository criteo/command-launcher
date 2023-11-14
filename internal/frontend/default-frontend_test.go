package frontend

import (
	"fmt"
	"testing"

	"github.com/criteo/command-launcher/internal/command"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func Test_FormatExamples(t *testing.T) {
	exampleText := formatExamples(nil)
	assert.Equal(t, "", exampleText)

	exampleText = formatExamples([]command.ExampleEntry{})
	assert.Equal(t, "", exampleText)

	exampleText = formatExamples([]command.ExampleEntry{
		{
			Scenario: "scenario 1",
			Command:  "test -O opt arg1 arg2",
		},
	})
	assert.Equal(t,
		`  # scenario 1
  test -O opt arg1 arg2
`, exampleText)
}

func Test_ParseFlagDefinition(t *testing.T) {
	name, short, desc, flagType, defaultValue := ParseFlagDefinition("test\t An test flag definition without short form")
	assert.Equal(t, "test", name)
	assert.Equal(t, "", short)
	assert.Equal(t, "An test flag definition without short form", desc)
	assert.Equal(t, "string", flagType)
	assert.Equal(t, "", defaultValue)

	name, short, desc, flagType, defaultValue = ParseFlagDefinition("test\t t \tAn test flag definition with short form")
	assert.Equal(t, "test", name)
	assert.Equal(t, "t", short)
	assert.Equal(t, "An test flag definition with short form", desc)
	assert.Equal(t, "string", flagType)
	assert.Equal(t, "", defaultValue)

	name, short, desc, flagType, defaultValue = ParseFlagDefinition("test\t t \tAn test flag definition with short form\t string\t ok")
	assert.Equal(t, "test", name)
	assert.Equal(t, "t", short)
	assert.Equal(t, "An test flag definition with short form", desc)
	assert.Equal(t, "string", flagType)
	assert.Equal(t, "ok", defaultValue)

	name, short, desc, flagType, defaultValue = ParseFlagDefinition("test")
	assert.Equal(t, "test", name)
	assert.Equal(t, "", short)
	assert.Equal(t, "", desc)
	assert.Equal(t, "string", flagType)
	assert.Equal(t, "", defaultValue)

	name, short, desc, flagType, defaultValue = ParseFlagDefinition("test\t t\tA test flag description\tbool")
	assert.Equal(t, "test", name)
	assert.Equal(t, "t", short)
	assert.Equal(t, "A test flag description", desc)
	assert.Equal(t, "bool", flagType)

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
	cmd.Flags().BoolP("flag-with-dash", "d", false, "description")

	envList, envTable, originalArgs, err := parseCmdArgsToEnv(cmd, []string{"--flag1", "v1", "-b", "v2", "-c", "--flag-with-dash", "arg1", "arg2"}, "CDT")
	assert.Nil(t, err)

	assert.True(t, findEnv(envList, "CDT_FLAG_FLAG1", "v1"))
	assert.True(t, findEnv(envList, "CDT_FLAG_FLAG2", "v2"))
	assert.True(t, findEnv(envList, "CDT_FLAG_FLAG3", "true"))
	assert.True(t, findEnv(envList, "CDT_FLAG_FLAG_WITH_DASH", "true"))

	assert.True(t, findEnv(envList, "CDT_ARG_1", "arg1"))
	assert.True(t, findEnv(envList, "CDT_ARG_2", "arg2"))

	assert.Equal(t, envTable["CDT_FLAG_FLAG1"], "v1")
	assert.Equal(t, envTable["CDT_FLAG_FLAG2"], "v2")
	assert.Equal(t, envTable["CDT_FLAG_FLAG3"], "true")
	assert.Equal(t, envTable["CDT_FLAG_FLAG_WITH_DASH"], "true")

	assert.Equal(t, envTable["CDT_ARG_1"], "arg1")
	assert.Equal(t, envTable["CDT_ARG_2"], "arg2")

	fmt.Println(originalArgs)
	assert.Equal(t, len(originalArgs), 8)
}

func findEnv(envTable []string, key string, value string) bool {
	for _, v := range envTable {
		if v == fmt.Sprintf("%s=%s", key, value) {
			return true
		}
	}
	return false
}
