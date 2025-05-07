package types_test

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

func TestLOCKING_ANDX_RANGE64_Marshal(t *testing.T) {
	tests := []struct {
		name     string
		input    types.LOCKING_ANDX_RANGE64
		expected []byte
	}{
		{
			name: "Basic Marshal",
			input: types.LOCKING_ANDX_RANGE64{
				PID:               0x1234,
				Pad:               0x5678,
				ByteOffsetHigh:    0x9ABCDEF0,
				ByteOffsetLow:     0x12345678,
				LengthInBytesHigh: 0xAABBCCDD,
				LengthInBytesLow:  0xEEFF0011,
			},
			expected: []byte{
				0x34, 0x12, // PID (little-endian)
				0x78, 0x56, // Pad (little-endian)
				0xF0, 0xDE, 0xBC, 0x9A, // ByteOffsetHigh (little-endian)
				0x78, 0x56, 0x34, 0x12, // ByteOffsetLow (little-endian)
				0xDD, 0xCC, 0xBB, 0xAA, // LengthInBytesHigh (little-endian)
				0x11, 0x00, 0xFF, 0xEE, // LengthInBytesLow (little-endian)
			},
		},
		{
			name: "Zero Values",
			input: types.LOCKING_ANDX_RANGE64{
				PID:               0,
				Pad:               0,
				ByteOffsetHigh:    0,
				ByteOffsetLow:     0,
				LengthInBytesHigh: 0,
				LengthInBytesLow:  0,
			},
			expected: []byte{
				0x00, 0x00, // PID
				0x00, 0x00, // Pad
				0x00, 0x00, 0x00, 0x00, // ByteOffsetHigh
				0x00, 0x00, 0x00, 0x00, // ByteOffsetLow
				0x00, 0x00, 0x00, 0x00, // LengthInBytesHigh
				0x00, 0x00, 0x00, 0x00, // LengthInBytesLow
			},
		},
		{
			name: "Maximum Values",
			input: types.LOCKING_ANDX_RANGE64{
				PID:               0xFFFF,
				Pad:               0xFFFF,
				ByteOffsetHigh:    0xFFFFFFFF,
				ByteOffsetLow:     0xFFFFFFFF,
				LengthInBytesHigh: 0xFFFFFFFF,
				LengthInBytesLow:  0xFFFFFFFF,
			},
			expected: []byte{
				0xFF, 0xFF, // PID
				0xFF, 0xFF, // Pad
				0xFF, 0xFF, 0xFF, 0xFF, // ByteOffsetHigh
				0xFF, 0xFF, 0xFF, 0xFF, // ByteOffsetLow
				0xFF, 0xFF, 0xFF, 0xFF, // LengthInBytesHigh
				0xFF, 0xFF, 0xFF, 0xFF, // LengthInBytesLow
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.input.Marshal()
			if err != nil {
				t.Errorf("Marshal() error = %v", err)
				return
			}
			if !bytes.Equal(result, tt.expected) {
				t.Errorf("Marshal() got = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestLOCKING_ANDX_RANGE64_Unmarshal(t *testing.T) {
	tests := []struct {
		name      string
		input     []byte
		expected  types.LOCKING_ANDX_RANGE64
		bytesRead int
		hasError  bool
	}{
		{
			name: "Basic Unmarshal",
			input: []byte{
				0x34, 0x12, // PID (little-endian)
				0x78, 0x56, // Pad (little-endian)
				0xF0, 0xDE, 0xBC, 0x9A, // ByteOffsetHigh (little-endian)
				0x78, 0x56, 0x34, 0x12, // ByteOffsetLow (little-endian)
				0xDD, 0xCC, 0xBB, 0xAA, // LengthInBytesHigh (little-endian)
				0x11, 0x00, 0xFF, 0xEE, // LengthInBytesLow (little-endian)
				0xFF, 0xFF, // Extra bytes that should be ignored
			},
			expected: types.LOCKING_ANDX_RANGE64{
				PID:               0x1234,
				Pad:               0x5678,
				ByteOffsetHigh:    0x9ABCDEF0,
				ByteOffsetLow:     0x12345678,
				LengthInBytesHigh: 0xAABBCCDD,
				LengthInBytesLow:  0xEEFF0011,
			},
			bytesRead: 20,
			hasError:  false,
		},
		{
			name: "Zero Values",
			input: []byte{
				0x00, 0x00, // PID
				0x00, 0x00, // Pad
				0x00, 0x00, 0x00, 0x00, // ByteOffsetHigh
				0x00, 0x00, 0x00, 0x00, // ByteOffsetLow
				0x00, 0x00, 0x00, 0x00, // LengthInBytesHigh
				0x00, 0x00, 0x00, 0x00, // LengthInBytesLow
			},
			expected: types.LOCKING_ANDX_RANGE64{
				PID:               0,
				Pad:               0,
				ByteOffsetHigh:    0,
				ByteOffsetLow:     0,
				LengthInBytesHigh: 0,
				LengthInBytesLow:  0,
			},
			bytesRead: 20,
			hasError:  false,
		},
		{
			name: "Maximum Values",
			input: []byte{
				0xFF, 0xFF, // PID
				0xFF, 0xFF, // Pad
				0xFF, 0xFF, 0xFF, 0xFF, // ByteOffsetHigh
				0xFF, 0xFF, 0xFF, 0xFF, // ByteOffsetLow
				0xFF, 0xFF, 0xFF, 0xFF, // LengthInBytesHigh
				0xFF, 0xFF, 0xFF, 0xFF, // LengthInBytesLow
			},
			expected: types.LOCKING_ANDX_RANGE64{
				PID:               0xFFFF,
				Pad:               0xFFFF,
				ByteOffsetHigh:    0xFFFFFFFF,
				ByteOffsetLow:     0xFFFFFFFF,
				LengthInBytesHigh: 0xFFFFFFFF,
				LengthInBytesLow:  0xFFFFFFFF,
			},
			bytesRead: 20,
			hasError:  false,
		},
		{
			name:      "Data Too Short",
			input:     []byte{0x34, 0x12, 0x78, 0x56, 0xF0, 0xDE, 0xBC, 0x9A, 0x78, 0x56, 0x34, 0x12, 0xDD, 0xCC, 0xBB}, // Only 15 bytes
			expected:  types.LOCKING_ANDX_RANGE64{},
			bytesRead: 0,
			hasError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result types.LOCKING_ANDX_RANGE64
			bytesRead, err := result.Unmarshal(tt.input)

			if tt.hasError && err == nil {
				t.Errorf("Unmarshal() expected error, got nil")
			}
			if !tt.hasError && err != nil {
				t.Errorf("Unmarshal() unexpected error: %v", err)
			}
			if !tt.hasError {
				if result.PID != tt.expected.PID {
					t.Errorf("Unmarshal() PID = %v, want %v", result.PID, tt.expected.PID)
				}
				if result.Pad != tt.expected.Pad {
					t.Errorf("Unmarshal() Pad = %v, want %v", result.Pad, tt.expected.Pad)
				}
				if result.ByteOffsetHigh != tt.expected.ByteOffsetHigh {
					t.Errorf("Unmarshal() ByteOffsetHigh = %v, want %v", result.ByteOffsetHigh, tt.expected.ByteOffsetHigh)
				}
				if result.ByteOffsetLow != tt.expected.ByteOffsetLow {
					t.Errorf("Unmarshal() ByteOffsetLow = %v, want %v", result.ByteOffsetLow, tt.expected.ByteOffsetLow)
				}
				if result.LengthInBytesHigh != tt.expected.LengthInBytesHigh {
					t.Errorf("Unmarshal() LengthInBytesHigh = %v, want %v", result.LengthInBytesHigh, tt.expected.LengthInBytesHigh)
				}
				if result.LengthInBytesLow != tt.expected.LengthInBytesLow {
					t.Errorf("Unmarshal() LengthInBytesLow = %v, want %v", result.LengthInBytesLow, tt.expected.LengthInBytesLow)
				}
				if bytesRead != tt.bytesRead {
					t.Errorf("Unmarshal() bytesRead = %v, want %v", bytesRead, tt.bytesRead)
				}
			}
		})
	}
}
