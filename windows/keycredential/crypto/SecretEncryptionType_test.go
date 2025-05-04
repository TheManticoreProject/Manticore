package crypto_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/keycredential/crypto"
)

func TestPrivateKeyEncryptionTypeString(t *testing.T) {
	tests := []struct {
		encryptionType crypto.PrivateKeyEncryptionType
		expected       string
	}{
		{crypto.NONE, "NONE"},
		{crypto.PasswordRC4, "PasswordRC4"},
		{crypto.PasswordRC2CBC, "PasswordRC2CBC"},
	}

	for _, test := range tests {
		t.Run(test.expected, func(t *testing.T) {
			result := test.encryptionType.String()
			if result != test.expected {
				t.Errorf("Expected %s, but got %s", test.expected, result)
			}
		})
	}
}
