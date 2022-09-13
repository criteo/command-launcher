package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ParseFlagDefinition(t *testing.T) {
	name, short, desc, flagType, defaultValue := parseFlagDefinition("test\t An test flag definition without short form")
	assert.Equal(t, "test", name)
	assert.Equal(t, "", short)
	assert.Equal(t, "An test flag definition without short form", desc)
	assert.Equal(t, "string", flagType)
	assert.Equal(t, "", defaultValue)

	name, short, desc, flagType, defaultValue = parseFlagDefinition("test\t t \tAn test flag definition with short form")
	assert.Equal(t, "test", name)
	assert.Equal(t, "t", short)
	assert.Equal(t, "An test flag definition with short form", desc)
	assert.Equal(t, "string", flagType)
	assert.Equal(t, "", defaultValue)

	name, short, desc, flagType, defaultValue = parseFlagDefinition("test\t t \tAn test flag definition with short form\t string\t ok")
	assert.Equal(t, "test", name)
	assert.Equal(t, "t", short)
	assert.Equal(t, "An test flag definition with short form", desc)
	assert.Equal(t, "string", flagType)
	assert.Equal(t, "ok", defaultValue)

	name, short, desc, flagType, defaultValue = parseFlagDefinition("test")
	assert.Equal(t, "test", name)
	assert.Equal(t, "", short)
	assert.Equal(t, "", desc)
	assert.Equal(t, "string", flagType)
	assert.Equal(t, "", defaultValue)

	name, short, desc, flagType, defaultValue = parseFlagDefinition("test\t t\tA test flag description\tbool")
	assert.Equal(t, "test", name)
	assert.Equal(t, "t", short)
	assert.Equal(t, "A test flag description", desc)
	assert.Equal(t, "bool", flagType)

}
