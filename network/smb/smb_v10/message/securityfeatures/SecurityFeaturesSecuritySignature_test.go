package securityfeatures_test

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/securityfeatures"
)

func TestNewSecurityFeaturesSecuritySignature(t *testing.T) {
	sf := securityfeatures.NewSecurityFeaturesSecuritySignature()
	if !bytes.Equal(sf.SecuritySignature[:], []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}) {
		t.Errorf("Expected SecuritySignature to be [0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00], got %v", sf.SecuritySignature)
	}
}

func TestSecurityFeaturesSecuritySignature_Marshal(t *testing.T) {
	sf := &securityfeatures.SecurityFeaturesSecuritySignature{
		SecuritySignature: [8]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
	}

	data, err := sf.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	expected := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	if len(data) != len(expected) {
		t.Fatalf("Expected data length %d, got %d", len(expected), len(data))
	}

	if !bytes.Equal(data, expected) {
		t.Errorf("Expected data %v, got %v", expected, data)
	}
}

func TestSecurityFeaturesSecuritySignature_Unmarshal(t *testing.T) {
	testCases := []struct {
		name        string
		data        []byte
		expected    [8]byte
		expectError bool
	}{
		{
			name:        "Valid data",
			data:        []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
			expected:    [8]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
			expectError: false,
		},
		{
			name:        "Invalid length",
			data:        []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07},
			expected:    [8]byte{},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sf := securityfeatures.NewSecurityFeaturesSecuritySignature()
			_, err := sf.Unmarshal(tc.data)

			if tc.expectError && err == nil {
				t.Errorf("Expected error but got none")
			} else if !tc.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tc.expectError {
				for i, b := range sf.SecuritySignature {
					if b != tc.expected[i] {
						t.Errorf("Byte %d: expected 0x%02x, got 0x%02x", i, tc.expected[i], b)
					}
				}
			}
		})
	}
}
