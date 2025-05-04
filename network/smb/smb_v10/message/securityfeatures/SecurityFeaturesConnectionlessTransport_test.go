package securityfeatures_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/securityfeatures"
)

func TestNewSecurityFeaturesConnectionlessTransport(t *testing.T) {
	sf := securityfeatures.NewSecurityFeaturesConnectionlessTransport()
	if sf.Key != 0x00000000 {
		t.Errorf("Expected Key to be 0x00000000, got 0x%08x", sf.Key)
	}
	if sf.CID != 0x0000 {
		t.Errorf("Expected CID to be 0x0000, got 0x%04x", sf.CID)
	}
	if sf.SequenceNumber != 0x0000 {
		t.Errorf("Expected SequenceNumber to be 0x0000, got 0x%04x", sf.SequenceNumber)
	}
}

func TestSecurityFeaturesConnectionlessTransport_Marshal(t *testing.T) {
	sf := &securityfeatures.SecurityFeaturesConnectionlessTransport{
		Key:            0x12345678,
		CID:            0xABCD,
		SequenceNumber: 0xEF01,
	}

	data, err := sf.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	expected := []byte{0x78, 0x56, 0x34, 0x12, 0xCD, 0xAB, 0x01, 0xEF}
	if len(data) != len(expected) {
		t.Fatalf("Expected data length %d, got %d", len(expected), len(data))
	}

	for i, b := range data {
		if b != expected[i] {
			t.Errorf("Byte %d: expected 0x%02x, got 0x%02x", i, expected[i], b)
		}
	}
}

func TestSecurityFeaturesConnectionlessTransport_Unmarshal(t *testing.T) {
	testCases := []struct {
		name          string
		data          []byte
		expectedKey   uint32
		expectedCID   uint16
		expectedSeqNo uint16
		expectError   bool
	}{
		{
			name:          "Valid data",
			data:          []byte{0x78, 0x56, 0x34, 0x12, 0xCD, 0xAB, 0x01, 0xEF},
			expectedKey:   0x12345678,
			expectedCID:   0xABCD,
			expectedSeqNo: 0xEF01,
			expectError:   false,
		},
		{
			name:        "Invalid length",
			data:        []byte{0x78, 0x56, 0x34, 0x12, 0xCD, 0xAB, 0x01},
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
			sf := &securityfeatures.SecurityFeaturesConnectionlessTransport{}
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

			if sf.Key != tc.expectedKey {
				t.Errorf("Expected Key to be 0x%08x, got 0x%08x", tc.expectedKey, sf.Key)
			}
			if sf.CID != tc.expectedCID {
				t.Errorf("Expected CID to be 0x%04x, got 0x%04x", tc.expectedCID, sf.CID)
			}
			if sf.SequenceNumber != tc.expectedSeqNo {
				t.Errorf("Expected SequenceNumber to be 0x%04x, got 0x%04x", tc.expectedSeqNo, sf.SequenceNumber)
			}
		})
	}
}

func TestSecurityFeaturesConnectionlessTransport_Interface(t *testing.T) {
	// Test that SecurityFeaturesConnectionlessTransport implements the SecurityFeatures interface
	var _ securityfeatures.SecurityFeatures = &securityfeatures.SecurityFeaturesConnectionlessTransport{}
}
