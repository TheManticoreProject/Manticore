package uuid_v2_test

import (
	"bytes"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/TheManticoreProject/Manticore/crypto/uuid/uuid_v2"
)

func TestUUIDv2Unmarshal(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{"ValidData", []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x27, 0x08, 0x10, 0x20, 0x30, 0x40, 0x50, 0x60, 0x70, 0x80}, false},
		{"InvalidLength", []byte{0x01, 0x02}, true},
		{"InvalidVersion", []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x20, 0x20, 0x30, 0x40, 0x50, 0x60, 0x70, 0x80}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var u uuid_v2.UUIDv2
			_, err := u.Unmarshal(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUUIDv2FromString(t *testing.T) {
	tests := []struct {
		name    string
		uuidStr string
		wantErr bool
	}{
		{"ValidString", "01020304-0506-2708-1020-304050607080", false},
		{"InvalidLength", "01020304-0506-2708-1020-30405060708", true},
		{"InvalidFormat", "01020304-0506-2708-1020-30405060708G", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var u uuid_v2.UUIDv2
			err := u.FromString(tt.uuidStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("FromString() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUUIDv2FromBytes(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{"ValidData", []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x27, 0x08, 0x10, 0x20, 0x30, 0x40, 0x50, 0x60, 0x70, 0x80}, false},
		{"InvalidLength", []byte{0x01, 0x02}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var u uuid_v2.UUIDv2
			err := u.FromBytes(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("FromBytes() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUUIDv2RoundTrip(t *testing.T) {
	tests := []struct {
		name    string
		uuidStr string
		wantErr bool
	}{
		{"UUIDv2_0_two_1", "00000000-0000-2000-0000-000000000000", false},
		{"UUIDv2_0_two_1", "10000000-0000-2000-0000-000000000000", false},
		{"UUIDv2_0_two_1", "01000000-0000-2000-0000-000000000000", false},
		{"UUIDv2_0_two_1", "00100000-0000-2000-0000-000000000000", false},
		{"UUIDv2_0_two_1", "00010000-0000-2000-0000-000000000000", false},
		{"UUIDv2_0_two_1", "00001000-0000-2000-0000-000000000000", false},
		{"UUIDv2_0_two_1", "00000100-0000-2000-0000-000000000000", false},
		{"UUIDv2_0_two_1", "00000010-0000-2000-0000-000000000000", false},
		{"UUIDv2_0_two_1", "00000001-0000-2000-0000-000000000000", false},
		{"UUIDv2_0_two_1", "00000000-1000-2000-0000-000000000000", false},
		{"UUIDv2_0_two_1", "00000000-0100-2000-0000-000000000000", false},
		{"UUIDv2_0_two_1", "00000000-0010-2000-0000-000000000000", false},
		{"UUIDv2_0_two_1", "00000000-0001-2000-0000-000000000000", false},
		{"UUIDv2_0_two_1", "00000000-0000-2000-0000-000000000000", false},
		{"UUIDv2_0_two_1", "00000000-0000-2100-0000-000000000000", false},
		{"UUIDv2_0_two_1", "00000000-0000-2010-0000-000000000000", false},
		{"UUIDv2_0_two_1", "00000000-0000-2001-0000-000000000000", false},
		{"UUIDv2_0_two_1", "00000000-0000-2000-1000-000000000000", false},
		{"UUIDv2_0_two_1", "00000000-0000-2000-0100-000000000000", false},
		{"UUIDv2_0_two_1", "00000000-0000-2000-0010-000000000000", false},
		{"UUIDv2_0_two_1", "00000000-0000-2000-0001-000000000000", false},
		{"UUIDv2_0_two_1", "00000000-0000-2000-0000-100000000000", false},
		{"UUIDv2_0_two_1", "00000000-0000-2000-0000-010000000000", false},
		{"UUIDv2_0_two_1", "00000000-0000-2000-0000-001000000000", false},
		{"UUIDv2_0_two_1", "00000000-0000-2000-0000-000100000000", false},
		{"UUIDv2_0_two_1", "00000000-0000-2000-0000-000010000000", false},
		{"UUIDv2_0_two_1", "00000000-0000-2000-0000-000001000000", false},
		{"UUIDv2_0_two_1", "00000000-0000-2000-0000-000000100000", false},
		{"UUIDv2_0_two_1", "00000000-0000-2000-0000-000000010000", false},
		{"UUIDv2_0_two_1", "00000000-0000-2000-0000-000000001000", false},
		{"UUIDv2_0_two_1", "00000000-0000-2000-0000-000000000100", false},
		{"UUIDv2_0_two_1", "00000000-0000-2000-0000-000000000010", false},
		{"UUIDv2_0_two_1", "00000000-0000-2000-0000-000000000001", false},

		{"timeHigh_and_clockSeq_0xAAA_0xBBB", "00000000-0000-2abc-1def-000000000000", false},

		{"UUIDv2_RFC4122_Example", "f81d4fae-7dec-21d0-a765-00a0c91e6bf6", false},
		{"UUIDv2_Example1", "9a3f1b40-140c-21ef-9c82-7b3eab180014", false},
		{"UUIDv2_Example2", "bd3c2722-140c-21ef-9646-4f98c463fb79", false},
		{"UUIDv2_Example3", "e4f5dc40-140c-21ef-84d3-138bfb4f0d3c", false},
		{"UUIDv2_Example4", "1d4e63b0-140d-21ef-927f-37c02b7d9f3a", false},
		{"UUIDv2_Example5", "3d9fae8e-140d-21ef-ae23-0f2adf4ebc2c", false},
		{"UUIDv2_Example6", "53c0295e-140d-21ef-813d-4f3f2a4dc499", false},
		{"UUIDv2_Example7", "692cbf20-140d-21ef-9f3f-1f9e308d2e00", false},
		{"UUIDv2_Example8", "84c31de2-140d-21ef-bc0f-2fa646f0d5d6", false},
		{"UUIDv2_Example9", "a0422e00-140d-21ef-aed9-173c88d9a19c", false},

		{"ValidUUIDv2", "01020304-0506-2708-8020-304050607080", false},
		{"ValidUUIDv2Alt", "01020304-0506-2708-1020-304050607080", false},
		{"ValidUUIDv2WithDashes", "6ba7b810-9dad-21d1-80b4-00c04fd430c8", false},
		{"ValidUUIDv2UpperCase", strings.ToUpper("01020304-0506-2708-8020-304050607080"), false},
		{"ValidUUIDv2MixedCase", strings.ToUpper("01020304-0506-2708-8020-304050607080"), false},
		{"InvalidVersion", "01020304-0506-4708-8020-304050607080", true},
		{"InvalidLength", "01020304-0506-2708-8020-30405060708", true},
		{"InvalidFormat", "01020304-0506-2708-8020-30405060708G", true},
		{"EmptyString", "", true},
		{"TooShort", "01020304", true},
		{"TooLong", "01020304-0506-2708-8020-304050607080-extra", true},
		{"InvalidCharacters", "01020304-0506-270Z-8020-304050607080", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// First, create a UUIDv2 from the string
			var u uuid_v2.UUIDv2
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

			// Convert back to string
			resultStr := u.String()

			// Normalize the input string for comparison (lowercase, no dashes)
			normalizedInput := strings.ReplaceAll(strings.ToLower(tt.uuidStr), "-", "")
			normalizedResult := strings.ReplaceAll(strings.ToLower(resultStr), "-", "")

			// Compare the normalized strings
			if normalizedResult != normalizedInput {
				t.Errorf("Round-trip conversion failed:\nInput--: %v\nOutput-: %v", tt.uuidStr, resultStr)
				t.Errorf("u.Time: 0x%016x\n", u.Time)
				t.Errorf("u.Clock: 0x%04x\n", u.Clock)
				t.Errorf("u.UUID.Data: %s\n", hex.EncodeToString(u.UUID.Data[:]))
			}

			// Verify the version is 2
			if u.UUID.Version != 2 {
				t.Errorf("Expected UUID version 2, got %d", u.UUID.Version)
			}
		})
	}
}

func TestUUIDv2ClockFromString(t *testing.T) {
	tests := []struct {
		name      string
		uuidStr   string
		wantClock uint8
		wantErr   bool
	}{
		{
			name:      "Standard UUIDv2",
			uuidStr:   "19c55c02-3406-21f0-9cd2-0242ac120002",
			wantClock: 0xc,
			wantErr:   false,
		},
		{
			name:      "Standard UUIDv2",
			uuidStr:   "861c3b82-3406-21f0-9fd2-0242ac120002",
			wantClock: 0xf,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var u uuid_v2.UUIDv2
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

			// Check if the clock sequence matches the expected value
			if u.Clock != tt.wantClock {
				t.Errorf("Clock sequence mismatch:\n\tgot  0x%04x (%d)\n\twant 0x%04x (%d)",
					u.Clock, u.Clock, tt.wantClock, tt.wantClock)
			}
		})
	}
}

func TestUUIDv2TimeFromString(t *testing.T) {
	tests := []struct {
		name           string
		uuidStr        string
		wantTimeString string
		wantTime       uint64
		wantErr        bool
	}{
		{
			name:           "Standard UUIDv2",
			uuidStr:        "19c55c02-3406-21f0-9cd2-0242ac120002",
			wantTimeString: "2025-05-18 16:35:25.529805.5 UTC", // 133920597255298050
			wantTime:       0x01f0340619c55c02,
			wantErr:        false,
		},
		{
			name:           "Standard UUIDv2",
			uuidStr:        "861c3b82-3406-21f0-9cd2-0242ac120002",
			wantTimeString: "2025-05-18 16:38:27.293069.2 UTC", // 133920599072930690
			wantTime:       0x01f03406861c3b82,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var u uuid_v2.UUIDv2
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

			// Check if the time field matches the expected value
			timeStr := u.GetTime().UTC().Format("2006-01-02 15:04:05.000000.0 UTC")
			if timeStr != tt.wantTimeString {
				t.Errorf("Time field mismatch:\n\tgot  %s\n\twant %s", timeStr, tt.wantTimeString)
			}

			if u.Time != tt.wantTime {
				t.Errorf("Time field mismatch:\n\tgot  0x%016x (%018d)\n\twant 0x%016x (%018d)", u.Time, u.Time, tt.wantTime, tt.wantTime)
			}
		})
	}
}

func TestUUIDv2MarshalUnmarshalPreservesTime(t *testing.T) {
	tests := []struct {
		name     string
		uuidStr  string
		timeUint uint64
		wantErr  bool
	}{
		{
			name: "Standard UUIDv2",
			// uuidStr:  "19c55c02-3406-21f0-9cd2-0242ac120002",
			timeUint: 133920597255298050,
			wantErr:  false,
		},
		{
			name: "Another UUIDv2",
			// uuidStr:  "861c3b82-3406-21f0-9cd2-0242ac120002",
			timeUint: 133920599072930690,
			wantErr:  false,
		},
		{
			name: "Another UUIDv2",
			// uuidStr:  "00000000-0000-2000-0000-000000000000",
			timeUint: 0,
			wantErr:  false,
		},
		{
			name: "Another UUIDv2",
			// uuidStr:  "00000000-0000-2000-0000-000000000000",
			timeUint: 0x0122334455667788,
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var u uuid_v2.UUIDv2
			u.Time = tt.timeUint

			// Verify marshaling and unmarshaling preserves the time value
			data, err := u.Marshal()
			if err != nil {
				t.Errorf("Marshal() error = %v", err)
				return
			}

			var u2 uuid_v2.UUIDv2
			_, err = u2.Unmarshal(data)
			if err != nil {
				t.Errorf("Unmarshal() error = %v", err)
				return
			}

			if u2.Time != u.Time {
				t.Errorf("Time field not preserved during marshal/unmarshal: original %#x, got %#x",
					u.Time, u2.Time)
			}

			// Also check that the string representation is preserved
			if u2.String() != u.String() {
				t.Errorf("String representation not preserved: original %s, got %s",
					u.String(), u2.String())
			}
		})
	}
}

func TestUUIDv2MarshalUnmarshalPreservesLocalDomain(t *testing.T) {
	tests := []struct {
		name            string
		wantLocalDomain uint8
		wantErr         bool
	}{
		{"LocalDomain 0", 0x0, false},
		{"LocalDomain 1", 0x1, false},
		{"LocalDomain 2", 0x2, false},
		{"LocalDomain 10", 0xa, false},
		{"LocalDomain 42", 0x2a, false},
		{"LocalDomain 100", 0x64, false},
		{"LocalDomain 127", 0x7f, false},
		{"LocalDomain 128", 0x80, false},
		{"LocalDomain 200", 0xc8, false},
		{"LocalDomain 255", 0xff, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var u uuid_v2.UUIDv2
			u.LocalDomain = tt.wantLocalDomain

			// Verify marshaling and unmarshaling preserves the local domain value
			data, err := u.Marshal()
			if err != nil {
				t.Errorf("Marshal() error = %v", err)
				return
			}

			var u2 uuid_v2.UUIDv2
			_, err = u2.Unmarshal(data)
			if err != nil {
				t.Errorf("Unmarshal() error = %v", err)
				return
			}

			if u2.LocalDomain != u.LocalDomain {
				t.Errorf("LocalDomain not preserved during marshal/unmarshal: original %#x, got %#x",
					u.LocalDomain, u2.LocalDomain)
			}

			// Also check that the string representation is preserved
			if u2.String() != u.String() {
				t.Errorf("String representation not preserved: original %s, got %s",
					u.String(), u2.String())
			}
		})
	}
}

func TestUUIDv2MarshalUnmarshalPreservesClock(t *testing.T) {
	tests := []struct {
		name      string
		wantClock uint8
		wantErr   bool
	}{
		{"Standard UUIDv2", 0x1, false},
		{"Standard UUIDv2", 0x2, false},
		{"Standard UUIDv2", 0x3, false},
		{"Standard UUIDv2", 0x4, false},
		{"Standard UUIDv2", 0x5, false},
		{"Standard UUIDv2", 0x6, false},
		{"Standard UUIDv2", 0x7, false},
		{"Standard UUIDv2", 0x8, false},
		{"Standard UUIDv2", 0x9, false},
		{"Standard UUIDv2", 0xa, false},
		{"Standard UUIDv2", 0xb, false},
		{"Standard UUIDv2", 0xc, false},
		{"Standard UUIDv2", 0xd, false},
		{"Standard UUIDv2", 0xe, false},
		{"Standard UUIDv2", 0xf, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var u uuid_v2.UUIDv2
			u.Clock = tt.wantClock

			// Verify marshaling and unmarshaling preserves the time value
			data, err := u.Marshal()
			if err != nil {
				t.Errorf("Marshal() error = %v", err)
				return
			}

			var u2 uuid_v2.UUIDv2
			_, err = u2.Unmarshal(data)
			if err != nil {
				t.Errorf("Unmarshal() error = %v", err)
				return
			}

			if u2.Clock != u.Clock {
				t.Errorf("Clock sequence not preserved during marshal/unmarshal: original %#x, got %#x",
					u.Clock, u2.Clock)
			}
		})
	}
}

func TestUUIDv2MarshalUnmarshalPreservesNodeID(t *testing.T) {
	tests := []struct {
		name       string
		wantNodeID [6]byte
		wantErr    bool
	}{
		{
			name:       "Standard UUIDv2",
			wantNodeID: [6]byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xab},
			wantErr:    false,
		},
		{
			name:       "Zero NodeID",
			wantNodeID: [6]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			wantErr:    false,
		},
		{
			name:       "MAC-like NodeID",
			wantNodeID: [6]byte{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff},
			wantErr:    false,
		},
		{
			name:       "Another NodeID",
			wantNodeID: [6]byte{0x12, 0x34, 0x56, 0x78, 0x90, 0xab},
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var u uuid_v2.UUIDv2
			copy(u.NodeID[:], tt.wantNodeID[:])

			// Verify marshaling and unmarshaling preserves the node ID
			data, err := u.Marshal()
			if err != nil {
				t.Errorf("Marshal() error = %v", err)
				return
			}

			var u2 uuid_v2.UUIDv2
			_, err = u2.Unmarshal(data)
			if err != nil {
				t.Errorf("Unmarshal() error = %v", err)
				return
			}

			if !bytes.Equal(u2.NodeID[:], u.NodeID[:]) {
				t.Errorf("NodeID not preserved during marshal/unmarshal: original %x, got %x",
					u.NodeID, u2.NodeID)
			}
		})
	}
}
