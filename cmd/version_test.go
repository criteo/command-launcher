package cmd

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GetVersion(t *testing.T) {
	v := getVersion("")
	assert.Equal(t, fmt.Sprintf("1.0.0, build dev-%s", os.Getenv("USER")), v, "invalid version")

	v = getVersion("123")
	assert.Equal(t, "1.0.0, build 123", v, "Invalid version")
}
