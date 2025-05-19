package uuid_v1_test

import (
	"bytes"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/TheManticoreProject/Manticore/crypto/uuid/uuid_v1"
)

func TestUUIDv1Unmarshal(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{"ValidData", []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x17, 0x08, 0x10, 0x20, 0x30, 0x40, 0x50, 0x60, 0x70, 0x80}, false},
		{"InvalidLength", []byte{0x01, 0x02}, true},
		{"InvalidVersion", []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x20, 0x20, 0x30, 0x40, 0x50, 0x60, 0x70, 0x80}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var u uuid_v1.UUIDv1
			_, err := u.Unmarshal(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUUIDv1FromString(t *testing.T) {
	tests := []struct {
		name    string
		uuidStr string
		wantErr bool
	}{
		{"ValidString", "01020304-0506-1708-1020-304050607080", false},
		{"InvalidLength", "01020304-0506-1708-1020-30405060708", true},
		{"InvalidFormat", "01020304-0506-1708-1020-30405060708G", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var u uuid_v1.UUIDv1
			err := u.FromString(tt.uuidStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("FromString() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUUIDv1FromBytes(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{"ValidData", []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x17, 0x08, 0x10, 0x20, 0x30, 0x40, 0x50, 0x60, 0x70, 0x80}, false},
		{"InvalidLength", []byte{0x01, 0x02}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var u uuid_v1.UUIDv1
			err := u.FromBytes(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("FromBytes() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUUIDv1RoundTrip(t *testing.T) {
	tests := []struct {
		name    string
		uuidStr string
		wantErr bool
	}{
		{"UUIDv1_0_two_1", "00000000-0000-1000-0000-000000000000", false},
		{"UUIDv1_0_two_1", "10000000-0000-1000-0000-000000000000", false},
		{"UUIDv1_0_two_1", "01000000-0000-1000-0000-000000000000", false},
		{"UUIDv1_0_two_1", "00100000-0000-1000-0000-000000000000", false},
		{"UUIDv1_0_two_1", "00010000-0000-1000-0000-000000000000", false},
		{"UUIDv1_0_two_1", "00001000-0000-1000-0000-000000000000", false},
		{"UUIDv1_0_two_1", "00000100-0000-1000-0000-000000000000", false},
		{"UUIDv1_0_two_1", "00000010-0000-1000-0000-000000000000", false},
		{"UUIDv1_0_two_1", "00000001-0000-1000-0000-000000000000", false},
		{"UUIDv1_0_two_1", "00000000-1000-1000-0000-000000000000", false},
		{"UUIDv1_0_two_1", "00000000-0100-1000-0000-000000000000", false},
		{"UUIDv1_0_two_1", "00000000-0010-1000-0000-000000000000", false},
		{"UUIDv1_0_two_1", "00000000-0001-1000-0000-000000000000", false},
		{"UUIDv1_0_two_1", "00000000-0000-1000-0000-000000000000", false},
		{"UUIDv1_0_two_1", "00000000-0000-1100-0000-000000000000", false},
		{"UUIDv1_0_two_1", "00000000-0000-1010-0000-000000000000", false},
		{"UUIDv1_0_two_1", "00000000-0000-1001-0000-000000000000", false},
		{"UUIDv1_0_two_1", "00000000-0000-1000-1000-000000000000", false},
		{"UUIDv1_0_two_1", "00000000-0000-1000-0100-000000000000", false},
		{"UUIDv1_0_two_1", "00000000-0000-1000-0010-000000000000", false},
		{"UUIDv1_0_two_1", "00000000-0000-1000-0001-000000000000", false},
		{"UUIDv1_0_two_1", "00000000-0000-1000-0000-100000000000", false},
		{"UUIDv1_0_two_1", "00000000-0000-1000-0000-010000000000", false},
		{"UUIDv1_0_two_1", "00000000-0000-1000-0000-001000000000", false},
		{"UUIDv1_0_two_1", "00000000-0000-1000-0000-000100000000", false},
		{"UUIDv1_0_two_1", "00000000-0000-1000-0000-000010000000", false},
		{"UUIDv1_0_two_1", "00000000-0000-1000-0000-000001000000", false},
		{"UUIDv1_0_two_1", "00000000-0000-1000-0000-000000100000", false},
		{"UUIDv1_0_two_1", "00000000-0000-1000-0000-000000010000", false},
		{"UUIDv1_0_two_1", "00000000-0000-1000-0000-000000001000", false},
		{"UUIDv1_0_two_1", "00000000-0000-1000-0000-000000000100", false},
		{"UUIDv1_0_two_1", "00000000-0000-1000-0000-000000000010", false},
		{"UUIDv1_0_two_1", "00000000-0000-1000-0000-000000000001", false},

		{"timeHigh_and_clockSeq_0xAAA_0xBBB", "00000000-0000-1abc-1def-000000000000", false},

		{"UUIDv1_RFC4122_Example", "f81d4fae-7dec-11d0-a765-00a0c91e6bf6", false},
		{"UUIDv1_Example1", "9a3f1b40-140c-11ef-9c82-7b3eab180014", false},
		{"UUIDv1_Example2", "bd3c2722-140c-11ef-9646-4f98c463fb79", false},
		{"UUIDv1_Example3", "e4f5dc40-140c-11ef-84d3-138bfb4f0d3c", false},
		{"UUIDv1_Example4", "1d4e63b0-140d-11ef-927f-37c02b7d9f3a", false},
		{"UUIDv1_Example5", "3d9fae8e-140d-11ef-ae23-0f2adf4ebc2c", false},
		{"UUIDv1_Example6", "53c0295e-140d-11ef-813d-4f3f2a4dc499", false},
		{"UUIDv1_Example7", "692cbf20-140d-11ef-9f3f-1f9e308d2e00", false},
		{"UUIDv1_Example8", "84c31de2-140d-11ef-bc0f-2fa646f0d5d6", false},
		{"UUIDv1_Example9", "a0422e00-140d-11ef-aed9-173c88d9a19c", false},

		{"ValidUUIDv1", "01020304-0506-1708-8020-304050607080", false},
		{"ValidUUIDv1Alt", "01020304-0506-1708-1020-304050607080", false},
		{"ValidUUIDv1WithDashes", "6ba7b810-9dad-11d1-80b4-00c04fd430c8", false},
		{"ValidUUIDv1UpperCase", strings.ToUpper("01020304-0506-1708-8020-304050607080"), false},
		{"ValidUUIDv1MixedCase", strings.ToUpper("01020304-0506-1708-8020-304050607080"), false},
		{"InvalidVersion", "01020304-0506-4708-8020-304050607080", true},
		{"InvalidLength", "01020304-0506-1708-8020-30405060708", true},
		{"InvalidFormat", "01020304-0506-1708-8020-30405060708G", true},
		{"EmptyString", "", true},
		{"TooShort", "01020304", true},
		{"TooLong", "01020304-0506-1708-8020-304050607080-extra", true},
		{"InvalidCharacters", "01020304-0506-170Z-8020-304050607080", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// First, create a UUIDv1 from the string
			var u uuid_v1.UUIDv1
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
				t.Errorf("u.ClockSeq: 0x%04x\n", u.ClockSeq)
				t.Errorf("u.UUID.Data: %s\n", hex.EncodeToString(u.UUID.Data[:]))
			}

			// Verify the version is 1
			if u.UUID.Version != 1 {
				t.Errorf("Expected UUID version 1, got %d", u.UUID.Version)
			}
		})
	}
}

func TestUUIDv1ClockSeqFromString(t *testing.T) {
	tests := []struct {
		name         string
		uuidStr      string
		wantClockSeq uint16
		wantErr      bool
	}{
		{
			name:         "Standard UUIDv1",
			uuidStr:      "19c55c02-3406-11f0-9cd2-0242ac120002",
			wantClockSeq: 0xcd2,
			wantErr:      false,
		},
		{
			name:         "Standard UUIDv1",
			uuidStr:      "861c3b82-3406-11f0-9cd2-0242ac120002",
			wantClockSeq: 0xcd2,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var u uuid_v1.UUIDv1
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
			if u.ClockSeq != tt.wantClockSeq {
				t.Errorf("Clock sequence mismatch:\n\tgot  0x%04x (%d)\n\twant 0x%04x (%d)",
					u.ClockSeq, u.ClockSeq, tt.wantClockSeq, tt.wantClockSeq)
			}
		})
	}
}

func TestUUIDv1TimeFromString(t *testing.T) {
	tests := []struct {
		name           string
		uuidStr        string
		wantTimeString string
		wantTime       uint64
		wantErr        bool
	}{
		{
			name:           "Standard UUIDv1",
			uuidStr:        "19c55c02-3406-11f0-9cd2-0242ac120002",
			wantTimeString: "2025-05-18 16:35:25.529805.5 UTC", // 133920597255298050
			wantTime:       0x01f0340619c55c02,
			wantErr:        false,
		},
		{
			name:           "Standard UUIDv1",
			uuidStr:        "861c3b82-3406-11f0-9cd2-0242ac120002",
			wantTimeString: "2025-05-18 16:38:27.293069.2 UTC", // 133920599072930690
			wantTime:       0x01f03406861c3b82,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var u uuid_v1.UUIDv1
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

func TestUUIDv1MarshalUnmarshalPreservesTime(t *testing.T) {
	tests := []struct {
		name     string
		uuidStr  string
		timeUint uint64
		wantErr  bool
	}{
		{
			name: "Standard UUIDv1",
			// uuidStr:  "19c55c02-3406-11f0-9cd2-0242ac120002",
			timeUint: 133920597255298050,
			wantErr:  false,
		},
		{
			name: "Another UUIDv1",
			// uuidStr:  "861c3b82-3406-11f0-9cd2-0242ac120002",
			timeUint: 133920599072930690,
			wantErr:  false,
		},
		{
			name: "Another UUIDv1",
			// uuidStr:  "00000000-0000-1000-0000-000000000000",
			timeUint: 0,
			wantErr:  false,
		},
		{
			name: "Another UUIDv1",
			// uuidStr:  "00000000-0000-1000-0000-000000000000",
			timeUint: 0x0122334455667788,
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var u uuid_v1.UUIDv1
			u.Time = tt.timeUint

			// Verify marshaling and unmarshaling preserves the time value
			data, err := u.Marshal()
			if err != nil {
				t.Errorf("Marshal() error = %v", err)
				return
			}

			var u2 uuid_v1.UUIDv1
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

func TestUUIDv1MarshalUnmarshalPreservesClockSeq(t *testing.T) {
	tests := []struct {
		name         string
		uuidStr      string
		wantClockSeq uint16
		wantErr      bool
	}{
		{
			name:         "Standard UUIDv1",
			wantClockSeq: 0xcd2,
			wantErr:      false,
		},
		{
			name:         "Another UUIDv1",
			wantClockSeq: 0,
			wantErr:      false,
		},
		{
			name:         "Another UUIDv1",
			wantClockSeq: 0x0123,
			wantErr:      false,
		},
		{
			name:         "Another UUIDv1",
			wantClockSeq: 0x0aaa,
			wantErr:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var u uuid_v1.UUIDv1
			u.ClockSeq = tt.wantClockSeq

			// Verify marshaling and unmarshaling preserves the time value
			data, err := u.Marshal()
			if err != nil {
				t.Errorf("Marshal() error = %v", err)
				return
			}

			var u2 uuid_v1.UUIDv1
			_, err = u2.Unmarshal(data)
			if err != nil {
				t.Errorf("Unmarshal() error = %v", err)
				return
			}

			if u2.ClockSeq != u.ClockSeq {
				t.Errorf("Clock sequence not preserved during marshal/unmarshal: original %#x, got %#x",
					u.ClockSeq, u2.ClockSeq)
			}
		})
	}
}

func TestUUIDv1MarshalUnmarshalPreservesNodeID(t *testing.T) {
	tests := []struct {
		name       string
		wantNodeID [6]byte
		wantErr    bool
	}{
		{
			name:       "Standard UUIDv1",
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
			var u uuid_v1.UUIDv1
			copy(u.NodeID[:], tt.wantNodeID[:])

			// Verify marshaling and unmarshaling preserves the node ID
			data, err := u.Marshal()
			if err != nil {
				t.Errorf("Marshal() error = %v", err)
				return
			}

			var u2 uuid_v1.UUIDv1
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
