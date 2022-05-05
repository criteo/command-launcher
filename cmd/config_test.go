package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrintableSettings(t *testing.T) {
	settings := map[string]interface{}{
		"key 1": "123",
		"key 0": "345",
		"abc":   "efg",
	}

	printable := printableSettingsInOrder(settings)

	assert.Equal(t, 3, len(printable))
	assert.Equal(t, "abc                                     : efg", printable[0])
	assert.Equal(t, "key 0                                   : 345", printable[1])
	assert.Equal(t, "key 1                                   : 123", printable[2])
}
