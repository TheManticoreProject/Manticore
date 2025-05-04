package gppp_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/crypto/gppp"
)

func Test_GPPPDecrypt(t *testing.T) {
	tests := []struct {
		encrypted string
		expected  string
	}{
		{"bdajdgpjZqolVYI3h2O2mp+JpxDuZd0xoi2M86z7JuI=", "Podalirius"},
	}

	for _, test := range tests {
		result, err := gppp.GPPPDecryptBase64(test.encrypted)
		if err != nil {
			t.Errorf("GPPPDecrypt(%q) failed: %v", test.encrypted, err)
		}

		if result != test.expected {
			t.Errorf("GPPPDecrypt(%q) = %q; expected %q", test.encrypted, result, test.expected)
		}
	}
}

func Test_GPPPEncrypt(t *testing.T) {
	tests := []struct {
		plaintext string
		expected  string
	}{
		{"Podalirius", "bdajdgpjZqolVYI3h2O2mp+JpxDuZd0xoi2M86z7JuI="},
	}

	for _, test := range tests {
		result, err := gppp.GPPPEncrypt(test.plaintext)
		if err != nil {
			t.Errorf("GPPPDecrypt(%q) failed: %v", test.plaintext, err)
		}

		if result != test.expected {
			t.Errorf("GPPPDecrypt(%q) = %q; expected %q", test.plaintext, result, test.expected)
		}
	}
}

func Test_GPPPInvolution(t *testing.T) {
	tests := []struct {
		plaintext string
	}{
		{"Podalirius"},
		{"password123"},
		{"SuperSecretPassword!"},
		{"Complex@P4ssw0rd#2023"},
		{""},
		{"This is a longer password with spaces and special chars !@#$%^&*()"},
		{"ÜberPasswörd"}, // Test with non-ASCII characters
	}

	for _, test := range tests {
		encrypted, err := gppp.GPPPEncrypt(test.plaintext)
		if err != nil {
			t.Errorf("GPPPEncrypt(%q) failed: %v", test.plaintext, err)
		}

		decrypted, err := gppp.GPPPDecryptBase64(encrypted)
		if err != nil {
			t.Errorf("GPPPDecryptBase64(%q) failed: %v", encrypted, err)
		}

		if decrypted != test.plaintext {
			t.Errorf("GPPPDecryptBase64(GPPPEncrypt(%q)) = %q; expected %q", test.plaintext, decrypted, test.plaintext)
		}
	}
}
