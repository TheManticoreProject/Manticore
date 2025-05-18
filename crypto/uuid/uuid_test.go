package uuid_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/crypto/uuid"
)

func TestMarshal(t *testing.T) {
	testCases := []struct {
		name     string
		input    [15]byte
		version  uint8
		variant  uint8
		expected string
		wantErr  bool
	}{
		{
			name:     "Placeholder UUID",
			input:    [15]byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff},
			version:  0,
			variant:  0,
			expected: "11223344-5566-0788-0799-aabbccddeeff",
			wantErr:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			u := &uuid.UUID{
				Version: tc.version,
				Variant: tc.variant,
				Data:    tc.input,
			}

			// Marshal the UUID
			data := u.String()

			// Check if the marshaled data matches the expected format
			if data != tc.expected {
				t.Errorf("Marshal produced incorrect data\nExpected-: %v\nGot------: %v", tc.expected, data)
			}

		})
	}
}

// TestFromString tests the FromString function
func TestFromString(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected uuid.UUID
		wantErr  bool
	}{
		{
			name:  "Placeholder UUID",
			input: "11223344-5566-0788-0799-aabbccddeeff",
			expected: uuid.UUID{
				Version: 0,
				Variant: 0,
			},
			wantErr: false,
		},
		{
			name:  "Valid UUID v1",
			input: "00000000-0000-1000-1000-000000000000",
			expected: uuid.UUID{
				Version: 1,
				Variant: 1,
			},
			wantErr: false,
		},
		{
			name:  "Valid UUID v4",
			input: "00000000-0000-4000-2000-000000000000",
			expected: uuid.UUID{
				Version: 4,
				Variant: 2,
			},
			wantErr: false,
		},
		{
			name:  "Valid UUID v5",
			input: "00000000-0000-5000-0000-000000000000",
			expected: uuid.UUID{
				Version: 5,
				Variant: 0,
			},
			wantErr: false,
		},
		{
			name:     "Invalid format - too short",
			input:    "00000000-0000-4000-8000-00000000",
			expected: uuid.UUID{},
			wantErr:  true,
		},
		{
			name:     "Invalid format - missing hyphens",
			input:    "000000000000400080000000000000000",
			expected: uuid.UUID{},
			wantErr:  true,
		},
		{
			name:     "Invalid format - non-hex characters",
			input:    "0000000g-0000-4000-8000-000000000000",
			expected: uuid.UUID{},
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			u := &uuid.UUID{}
			err := u.FromString(tc.input)

			// Check error status
			if (err != nil) != tc.wantErr {
				t.Errorf("FromString() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			// Skip further checks if we expected an error
			if tc.wantErr {
				return
			}

			// Check if the parsed UUID matches the expected UUID
			if u.Version != tc.expected.Version || u.Variant != tc.expected.Variant {
				t.Errorf("FromString() got = %v, want %v", u, tc.expected)
			}

			// Verify round-trip conversion
			if u.String() != tc.input {
				t.Errorf("Round-trip conversion failed: got %v, want %v", u.String(), tc.input)
			}
		})
	}
}

func TestUUIDStringRoundTrip(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		version uint8
		variant uint8
		wantErr bool
	}{
		{
			name:    "Valid UUID v4",
			input:   "11223344-5566-0788-0799-aabbccddeeff",
			version: 0,
			variant: 0,
			wantErr: false,
		},
		{
			name:    "Valid UUID v4",
			input:   "f47ac10b-58cc-4372-a567-0e02b2c3d479",
			version: 4,
			variant: 0xa,
			wantErr: false,
		},
		{
			name:    "Valid UUID v1",
			input:   "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
			version: 1,
			variant: 8,
			wantErr: false,
		},
		{
			name:    "Valid UUID with all zeros",
			input:   "00000000-0000-0000-0000-000000000000",
			version: 0,
			variant: 0,
			wantErr: false,
		},
		{
			name:    "Valid UUID with all fs",
			input:   "ffffffff-ffff-ffff-ffff-ffffffffffff",
			version: 0xf,
			variant: 0xf,
			wantErr: false,
		},
		{
			name:    "Invalid format - too short",
			input:   "f47ac10b-58cc-4372-a567-0e02b2c3d4",
			version: 0,
			variant: 0,
			wantErr: true,
		},
		{
			name:    "Invalid format - non-hex characters",
			input:   "g47ac10b-58cc-4372-a567-0e02b2c3d479",
			version: 0,
			variant: 0,
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Parse the input string to UUID
			u := &uuid.UUID{}
			err := u.FromString(tc.input)

			// Check error status
			if (err != nil) != tc.wantErr {
				t.Errorf("FromString() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			if u.Version != tc.version {
				t.Errorf("FromString() uuid.Version got = %v, want %v", u.Version, tc.version)
			}

			if u.Variant != tc.variant {
				t.Errorf("FromString() uuid.Variant got = %v, want %v", u.Variant, tc.variant)
			}

			// Skip further checks if we expected an error
			if tc.wantErr {
				return
			}

			// Convert back to string and verify round-trip
			result := u.String()
			if result != tc.input {
				t.Errorf("String round-trip failed: \nGot--: %v\nWant-: %v", result, tc.input)
			}
		})
	}
}
