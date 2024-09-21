package kv

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSecretsManagerGet(t *testing.T) {
	client := NewDummySecretsManagerClient(map[string]string{
		"key":            "key::value",
		"path/to/secret": "{\"key\":\"path/to/secret::key::value\", \"key.multiple.levels\":\"path/to/secret::key.multiple.levels::value\"}",
	})
	secretsManager := NewSecretsManagerWithClient(client)

	tests := []struct {
		name     string
		key      string
		expected string
		err      bool
	}{

		{
			name:     "Key",
			key:      "key",
			expected: "key::value",
			err:      false,
		},
		{
			name:     "KeyNotFound",
			key:      "notfound",
			expected: "",
			err:      true,
		},
		{
			name:     "PathAndKey",
			key:      "path/to/secret.key",
			expected: "path/to/secret::key::value",
			err:      false,
		},
		{
			name:     "PathAndKeyInvalidJSON",
			key:      "key.invalid",
			expected: "",
			err:      true,
		},
		{
			name:     "PathAndKeyMultipleLevels",
			key:      "path/to/secret.key.multiple.levels",
			expected: "path/to/secret::key.multiple.levels::value",
			err:      false,
		},
		{
			name:     "PathAndKeyNotFound",
			key:      "path/to/secret.notfound",
			expected: "",
			err:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := secretsManager.Get(tt.key)
			if tt.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expected, actual)
		})
	}
}
