package crypto_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/keycredentiallink/crypto"
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

// A value shorter than the 4-byte field must be rejected with an error rather
// than slicing past the end of the input.
func TestSecretEncryptionTypeUnmarshalShortInput(t *testing.T) {
	for _, input := range [][]byte{nil, {}, {0x10}, {0x10, 0x00, 0x00}} {
		var set crypto.SecretEncryptionType
		if _, err := set.Unmarshal(input); err == nil {
			t.Errorf("Unmarshal(%x) returned no error", input)
		}
	}
	var set crypto.SecretEncryptionType
	n, err := set.Unmarshal([]byte{0x13, 0x00, 0x00, 0x00})
	if err != nil || n != 4 || set.Value != crypto.SecretEncryptionType_DatabaseAES {
		t.Errorf("Unmarshal(13000000) = %d, %v, Value %#x; want 4, nil, 0x13", n, err, set.Value)
	}
}
