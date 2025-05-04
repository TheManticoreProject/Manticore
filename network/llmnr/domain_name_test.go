package llmnr

import (
	"bytes"
	"testing"
)

func TestValidateDomainName(t *testing.T) {
	tests := []struct {
		name     string
		expected error
	}{
		{"example.com", nil},
		{"a-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-long-domain-name-that-exceeds-the-maximum-allowed-length-for-a-domain-name.com", ErrNameTooLong},
		{"valid-domain.local", nil},
		{"another.valid-domain.local", nil},
		{"invalid_domain_with_underscores.com", nil}, // Assuming underscores are allowed in this context
		{"", nil}, // Empty domain name should be valid
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateDomainName(test.name)
			if err != test.expected {
				t.Errorf("ValidateDomainName = %v; want %v", err, test.expected)
			}
		})
	}
}

func TestEncodeDomainName(t *testing.T) {
	tests := []struct {
		name     string
		expected []byte
	}{
		{"example.com", []byte{7, 'e', 'x', 'a', 'm', 'p', 'l', 'e', 3, 'c', 'o', 'm', 0}},
	}

	for _, test := range tests {
		t.Run("EncodeDomainName", func(t *testing.T) {
			encoded, err := EncodeDomainName(test.name)
			if err != nil {
				t.Fatalf("failed to encode domain name: %v", err)
			}
			if !bytes.Equal(encoded, test.expected) {
				t.Errorf("EncodeDomainName = %v; want %v", encoded, test.expected)
			}
		})
	}
}

func TestDecodeDomainName(t *testing.T) {
	tests := []struct {
		data     []byte
		expected string
	}{
		{
			data:     []byte{7, 'e', 'x', 'a', 'm', 'p', 'l', 'e', 3, 'c', 'o', 'm', 0},
			expected: "example.com",
		},
	}

	for _, test := range tests {
		t.Run("DecodeDomainName", func(t *testing.T) {
			var offset int
			decoded, _, err := DecodeDomainName(test.data, offset)
			if err != nil {
				t.Fatalf("failed to decode domain name: %v", err)
			}
			if decoded != test.expected {
				t.Errorf("DecodeDomainName = %v; want %v", decoded, test.expected)
			}
		})
	}
}
