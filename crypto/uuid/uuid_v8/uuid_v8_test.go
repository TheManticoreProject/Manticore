package uuid_v8_test

import (
	"encoding/hex"
	"testing"

	"github.com/TheManticoreProject/Manticore/crypto/uuid/uuid_v8"
)

func TestUUIDv8(t *testing.T) {
	tests := []struct {
		name     string
		uuidStr  string
		wantData string
		wantErr  bool
	}{
		{
			name:     "Basic UUID v8",
			uuidStr:  "12345678-9abc-8def-1234-56789abcdef1",
			wantData: "123456789abcdef23456789abcdef1",
			wantErr:  false,
		},
		{
			name:     "Basic UUID v8",
			uuidStr:  "8d9aeee5-d9ad-8934-84f4-ac533183424d",
			wantData: "8d9aeee5d9ad9344f4ac533183424d",
			wantErr:  false,
		},
		{
			name:     "Another UUID v8",
			uuidStr:  "c0819443-a39c-8e47-a949-303520cf9661",
			wantData: "c0819443a39ce47949303520cf9661",
			wantErr:  false,
		},
		{
			name:     "Third UUID v8",
			uuidStr:  "17aae0f3-3230-84cf-ad4c-ca7b64fecff6",
			wantData: "17aae0f332304cfd4cca7b64fecff6",
			wantErr:  false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var u uuid_v8.UUIDv8

			err := u.FromString(test.uuidStr)

			// Check if error matches expectation
			if (err != nil) != test.wantErr {
				t.Errorf("FromString() error = %v, wantErr %v", err, test.wantErr)
				return
			}

			// Skip further checks if we expected an error
			if test.wantErr {
				return
			}

			hexData := hex.EncodeToString(u.Data[:])

			// Check if the data matches the expected value
			if hexData != test.wantData {
				t.Errorf("Data mismatch:\n\tgot  %s\n\twant %s", hexData, test.wantData)
			}

			// Test marshaling and unmarshaling
			marshalledData, err := u.Marshal()
			if err != nil {
				t.Errorf("Marshal() error = %v", err)
				return
			}

			var u2 uuid_v8.UUIDv8
			_, err = u2.Unmarshal(marshalledData)
			if err != nil {
				t.Errorf("Unmarshal() error = %v", err)
				return
			}

			// Check that the data is preserved
			if u2.Data != u.Data {
				t.Errorf("Data not preserved during marshal/unmarshal: original %x, got %x",
					u.Data, u2.Data)
			}

			// Check that the string representation is preserved
			if u2.String() != u.String() {
				t.Errorf("String representation not preserved: original %s, got %s",
					u.String(), u2.String())
			}

			// Verify the string matches the input
			if u.String() != test.uuidStr {
				t.Errorf("UUIDv8.String() \n\tgot  %v\n\twant %v", u.String(), test.uuidStr)
			}
		})
	}
}

func TestUUIDv8FromStringRoundTrip(t *testing.T) {
	tests := []struct {
		name    string
		uuidStr string
		wantErr bool
	}{
		{"Valid UUIDv8", "00000000-0000-8000-0000-000000000000", false},
		{"Valid UUIDv8", "10000000-0000-8000-0000-000000000000", false},
		{"Valid UUIDv8", "01000000-0000-8000-0000-000000000000", false},
		{"Valid UUIDv8", "00100000-0000-8000-0000-000000000000", false},
		{"Valid UUIDv8", "00010000-0000-8000-0000-000000000000", false},
		{"Valid UUIDv8", "00001000-0000-8000-0000-000000000000", false},
		{"Valid UUIDv8", "00000100-0000-8000-0000-000000000000", false},
		{"Valid UUIDv8", "00000010-0000-8000-0000-000000000000", false},
		{"Valid UUIDv8", "00000001-0000-8000-0000-000000000000", false},
		{"Valid UUIDv8", "00000000-1000-8000-0000-000000000000", false},
		{"Valid UUIDv8", "00000000-0100-8000-0000-000000000000", false},
		{"Valid UUIDv8", "00000000-0010-8000-0000-000000000000", false},
		{"Valid UUIDv8", "00000000-0001-8000-0000-000000000000", false},
		{"Valid UUIDv8", "00000000-0000-8000-0000-000000000000", false},
		{"Valid UUIDv8", "00000000-0000-8100-0000-000000000000", false},
		{"Valid UUIDv8", "00000000-0000-8010-0000-000000000000", false},
		{"Valid UUIDv8", "00000000-0000-8001-0000-000000000000", false},
		{"Valid UUIDv8", "00000000-0000-8000-1000-000000000000", false},
		{"Valid UUIDv8", "00000000-0000-8000-0100-000000000000", false},
		{"Valid UUIDv8", "00000000-0000-8000-0010-000000000000", false},
		{"Valid UUIDv8", "00000000-0000-8000-0001-000000000000", false},
		{"Valid UUIDv8", "00000000-0000-8000-0000-100000000000", false},
		{"Valid UUIDv8", "00000000-0000-8000-0000-010000000000", false},
		{"Valid UUIDv8", "00000000-0000-8000-0000-001000000000", false},
		{"Valid UUIDv8", "00000000-0000-8000-0000-000100000000", false},
		{"Valid UUIDv8", "00000000-0000-8000-0000-000010000000", false},
		{"Valid UUIDv8", "00000000-0000-8000-0000-000001000000", false},
		{"Valid UUIDv8", "00000000-0000-8000-0000-000000100000", false},
		{"Valid UUIDv8", "00000000-0000-8000-0000-000000010000", false},
		{"Valid UUIDv8", "00000000-0000-8000-0000-000000001000", false},
		{"Valid UUIDv8", "00000000-0000-8000-0000-000000000100", false},
		{"Valid UUIDv8", "00000000-0000-8000-0000-000000000010", false},
		{"Valid UUIDv8", "00000000-0000-8000-0000-000000000001", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var u uuid_v8.UUIDv8
			err := u.FromString(tt.uuidStr)

			// Check if error matches expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("FromString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Skip further checks if we expected an error
			if tt.wantErr {
				return
			}

			// Verify the string matches the input
			if u.String() != tt.uuidStr {
				t.Errorf("UUIDv8.String() \n\tgot  %v\n\twant %v", u.String(), tt.uuidStr)
			}

			// Test marshaling and unmarshaling
			marshalledData, err := u.Marshal()
			if err != nil {
				t.Errorf("Marshal() error = %v", err)
				return
			}

			var u2 uuid_v8.UUIDv8
			_, err = u2.Unmarshal(marshalledData)
			if err != nil {
				t.Errorf("Unmarshal() error = %v", err)
				return
			}

			// Check that the string representation is preserved
			if u2.String() != tt.uuidStr {
				t.Errorf("String representation not preserved: original %s, got %s",
					tt.uuidStr, u2.String())
			}
		})
	}
}
