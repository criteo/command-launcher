package cmd

import (
	"testing"

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
