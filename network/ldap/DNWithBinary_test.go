package ldap_test

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/ldap"
)

func TestUnmarshal(t *testing.T) {
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
			expectedBin: []byte{0x48, 0x65, 0x6c, 0x6c, 0x6f}, // "Hello" in hex
		},
		{
			name:        "Invalid size (size field does not match binary length)",
			rawBytes:    []byte("B:8:48656c6c6f:CN=John Doe,OU=Users,DC=example,DC=com"),
			expectError: true, // Only 5 bytes (10 hex characters), so size should be 10, not 8
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
		{
			name:        "Empty input",
			rawBytes:    []byte(""),
			expectError: true,
		},
		{
			name:        "Non-numeric size",
			rawBytes:    []byte("B:XX:48656c6c6f:CN=Name"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &ldap.DNWithBinary{}
			bytesRead, err := d.Unmarshal(tt.rawBytes)
			if tt.expectError {
				if err == nil {
					t.Errorf("Unmarshal() expected error but got none")
				}
				return
			} else if err != nil {
				t.Errorf("Unmarshal() unexpected error: %v", err)
				return
			}
			if bytesRead != len(tt.rawBytes) {
				t.Errorf("Unmarshal() bytesRead = %v, want %v", bytesRead, len(tt.rawBytes))
			}
			if d.DistinguishedName != tt.expectedDN {
				t.Errorf("Expected DistinguishedName = %v, got %v", tt.expectedDN, d.DistinguishedName)
			}
			if !bytes.Equal(d.BinaryData, tt.expectedBin) {
				t.Errorf("Expected BinaryData = %v, got %v (hex=%s)", tt.expectedBin, d.BinaryData, hex.EncodeToString(d.BinaryData))
			}
		})
	}
}
