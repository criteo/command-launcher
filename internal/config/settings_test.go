package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetSelfUpdatePolicyConfig(t *testing.T) {
	testCases := []struct {
		value         string
		expectedError bool
		description   string
	}{
		{
			value:         "exact_match",
			expectedError: false,
			description:   "Valid exact_match policy should succeed",
		},
		{
			value:         "only_newer_version",
			expectedError: false,
			description:   "Valid only_newer_version policy should succeed",
		},
		{
			value:         "invalid_policy",
			expectedError: true,
			description:   "Invalid policy should fail with error",
		},
		{
			value:         "",
			expectedError: true,
			description:   "Empty policy should fail with error",
		},
		{
			value:         "EXACT_MATCH",
			expectedError: true,
			description:   "Case sensitive - uppercase should fail",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			err := setSelfUpdatePolicyConfig(tc.value)
			if tc.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid value for self-update policy")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
