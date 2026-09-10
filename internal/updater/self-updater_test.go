package updater

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsNewerVersion(t *testing.T) {
	assert.True(t, isNewerVersion("1.16.0", "1.15.2"))
	assert.True(t, isNewerVersion("1.10.0", "1.9.0"))
	assert.False(t, isNewerVersion("1.15.2", "1.16.0"))
	assert.False(t, isNewerVersion("1.16.0", "1.16.0"))
	assert.True(t, isNewerVersion("1.16", "1.15.9"))
	assert.False(t, isNewerVersion("1.16.0", "1.16"))
	assert.True(t, isNewerVersion("2.0.0", "1.99.99"))
}

func TestIsNewerVersionFallbackOnNonNumericSuffix(t *testing.T) {
	assert.True(t, isNewerVersion("1.16.0-beta", "1.15.2"))
	assert.False(t, isNewerVersion("1.16.0", "1.16.0"))
	assert.True(t, isNewerVersion("1.16.0", "1.16.0-beta"))
	assert.False(t, isNewerVersion("1.16.0-beta", "1.16.0-beta"))
}
