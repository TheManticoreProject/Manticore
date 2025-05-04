package keycredentiallink

import (
	"bytes"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name        string
		rawBytes    []byte
		expectError bool
		expectedDN  string
		expectedBin []byte
	}{
		{
			name:        "Valid input",
			rawBytes:    []byte("B:10:48656c6c6f:CN=John Doe,OU=Users,DC=example,DC=com"),
			expectError: false,
			expectedDN:  "CN=John Doe,OU=Users,DC=example,DC=com",
			expectedBin: []byte("Hello"),
		},
		{
			name:        "Invalid size",
			rawBytes:    []byte("B:5:48656c6c6f:CN=John Doe,OU=Users,DC=example,DC=com"),
			expectError: true,
		},
		{
			name:        "Invalid hex string",
			rawBytes:    []byte("B:10:ZZZZ:CN=John Doe,OU=Users,DC=example,DC=com"),
			expectError: true,
		},
		{
			name:        "Invalid parts count",
			rawBytes:    []byte("B:10:48656c6c6f"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DNWithBinary{}
			err := d.Parse(tt.rawBytes)
			if (err != nil) != tt.expectError {
				t.Errorf("Parse() error = %v, expectError %v", err, tt.expectError)
				return
			}
			if !tt.expectError {
				if d.DistinguishedName != tt.expectedDN {
					t.Errorf("Expected DistinguishedName = %v, got %v", tt.expectedDN, d.DistinguishedName)
				}
				if !bytes.Equal(d.BinaryData, tt.expectedBin) {
					t.Errorf("Expected BinaryData = %v, got %v", tt.expectedBin, d.BinaryData)
				}
			}
		})
	}
}
