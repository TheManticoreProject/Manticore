package securityfeatures_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/securityfeatures"
)

func TestNewSecurityFeaturesReserved(t *testing.T) {
	sf := securityfeatures.NewSecurityFeaturesReserved()
	for i, b := range sf.Reserved {
		if b != 0x00 {
			t.Errorf("Expected Reserved[%d] to be 0x00, got 0x%02x", i, b)
		}
	}
}

func TestSecurityFeaturesReserved_Marshal(t *testing.T) {
	sf := &securityfeatures.SecurityFeaturesReserved{
		Reserved: [8]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
	}

	data, err := sf.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	expected := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	if len(data) != len(expected) {
		t.Fatalf("Expected data length %d, got %d", len(expected), len(data))
	}

	for i, b := range data {
		if b != expected[i] {
			t.Errorf("Byte %d: expected 0x%02x, got 0x%02x", i, expected[i], b)
		}
	}
}

func TestSecurityFeaturesReserved_Unmarshal(t *testing.T) {
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
			expectError: true,
		},
		{
			name:        "Empty data",
			data:        []byte{},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sf := &securityfeatures.SecurityFeaturesReserved{}
			_, err := sf.Unmarshal(tc.data)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			for i, b := range sf.Reserved {
				if b != tc.expected[i] {
					t.Errorf("Byte %d: expected 0x%02x, got 0x%02x", i, tc.expected[i], b)
				}
			}
		})
	}
}

func TestSecurityFeaturesReserved_Interface(t *testing.T) {
	// Test that SecurityFeaturesReserved implements the SecurityFeatures interface
	var _ securityfeatures.SecurityFeatures = &securityfeatures.SecurityFeaturesReserved{}
}
