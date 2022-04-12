package helper

import (
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDarwinDnsResolve(t *testing.T) {
	if runtime.GOOS == "darwin" {
		ip, resolved := DarwinDnsResolve("https://go.dev")
		assert.True(t, resolved)
		assert.NotEmpty(t, ip)

		parts := strings.Split(ip, ".")
		assert.Equal(t, 4, len(parts))
		for i := 0; i < 4; i++ {
			d, err := strconv.Atoi(parts[i])
			assert.Nil(t, err)
			assert.True(t, d <= 255)
		}
	}
}
