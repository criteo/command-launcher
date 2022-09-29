package consent

import (
	"testing"

	"github.com/criteo/command-launcher/internal/context"
	"github.com/stretchr/testify/assert"
)

func TestAccessConsents(t *testing.T) {
	// TODO: we shouldn't let the secret lib depends on the context
	context.InitContext("test-vault", "1.0.0", "1")

	err := saveCmdConsents("dev-group", "test-cmd", []string{
		"USERNAME", "PASSWORD", "LOG_LEVEL",
	}, 30)

	assert.Nil(t, err)

	consent, err := getCmdConsents("dev-group", "test-cmd")

	assert.Nil(t, err)
	assert.NotNil(t, consent)
	assert.Equal(t, 3, len(consent.Consents))
	assert.Equal(t, "USERNAME", consent.Consents[0])
	assert.Equal(t, "PASSWORD", consent.Consents[1])
	assert.Equal(t, "LOG_LEVEL", consent.Consents[2])
}

func TestEmptyConsents(t *testing.T) {
	consent, err := GetConsents("non-exist-group", "no-exist-cmd", []string{}, true)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(consent))
}
