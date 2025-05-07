package types_test

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

func TestLOCKING_ANDX_RANGE32_Marshal(t *testing.T) {
	tests := []struct {
		name     string
		input    types.LOCKING_ANDX_RANGE32
		expected []byte
	}{
		{
			name: "Basic Marshal",
			input: types.LOCKING_ANDX_RANGE32{
				PID:           0x1234,
				ByteOffset:    0x56789ABC,
				LengthInBytes: 0xDEF01234,
			},
			expected: []byte{
				0x34, 0x12, // PID (little-endian)
				0xBC, 0x9A, 0x78, 0x56, // ByteOffset (little-endian)
				0x34, 0x12, 0xF0, 0xDE, // LengthInBytes (little-endian)
			},
		},
		{
			name: "Zero Values",
			input: types.LOCKING_ANDX_RANGE32{
				PID:           0,
				ByteOffset:    0,
				LengthInBytes: 0,
			},
			expected: []byte{
				0x00, 0x00, // PID
				0x00, 0x00, 0x00, 0x00, // ByteOffset
				0x00, 0x00, 0x00, 0x00, // LengthInBytes
			},
		},
		{
			name: "Maximum Values",
			input: types.LOCKING_ANDX_RANGE32{
				PID:           0xFFFF,
				ByteOffset:    0xFFFFFFFF,
				LengthInBytes: 0xFFFFFFFF,
			},
			expected: []byte{
				0xFF, 0xFF, // PID
				0xFF, 0xFF, 0xFF, 0xFF, // ByteOffset
				0xFF, 0xFF, 0xFF, 0xFF, // LengthInBytes
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

func TestLOCKING_ANDX_RANGE32_Unmarshal(t *testing.T) {
	tests := []struct {
		name      string
		input     []byte
		expected  types.LOCKING_ANDX_RANGE32
		bytesRead int
		hasError  bool
	}{
		{
			name: "Basic Unmarshal",
			input: []byte{
				0x34, 0x12, // PID (little-endian)
				0xBC, 0x9A, 0x78, 0x56, // ByteOffset (little-endian)
				0x34, 0x12, 0xF0, 0xDE, // LengthInBytes (little-endian)
				0xFF, 0xFF, // Extra bytes that should be ignored
			},
			expected: types.LOCKING_ANDX_RANGE32{
				PID:           0x1234,
				ByteOffset:    0x56789ABC,
				LengthInBytes: 0xDEF01234,
			},
			bytesRead: 10,
			hasError:  false,
		},
		{
			name: "Zero Values",
			input: []byte{
				0x00, 0x00, // PID
				0x00, 0x00, 0x00, 0x00, // ByteOffset
				0x00, 0x00, 0x00, 0x00, // LengthInBytes
			},
			expected: types.LOCKING_ANDX_RANGE32{
				PID:           0,
				ByteOffset:    0,
				LengthInBytes: 0,
			},
			bytesRead: 10,
			hasError:  false,
		},
		{
			name: "Maximum Values",
			input: []byte{
				0xFF, 0xFF, // PID
				0xFF, 0xFF, 0xFF, 0xFF, // ByteOffset
				0xFF, 0xFF, 0xFF, 0xFF, // LengthInBytes
			},
			expected: types.LOCKING_ANDX_RANGE32{
				PID:           0xFFFF,
				ByteOffset:    0xFFFFFFFF,
				LengthInBytes: 0xFFFFFFFF,
			},
			bytesRead: 10,
			hasError:  false,
		},
		{
			name:      "Data Too Short",
			input:     []byte{0x34, 0x12, 0xBC, 0x9A, 0x78, 0x56, 0x34, 0x12}, // Only 8 bytes
			expected:  types.LOCKING_ANDX_RANGE32{},
			bytesRead: 0,
			hasError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result types.LOCKING_ANDX_RANGE32
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
				if result.ByteOffset != tt.expected.ByteOffset {
					t.Errorf("Unmarshal() ByteOffset = %v, want %v", result.ByteOffset, tt.expected.ByteOffset)
				}
				if result.LengthInBytes != tt.expected.LengthInBytes {
					t.Errorf("Unmarshal() LengthInBytes = %v, want %v", result.LengthInBytes, tt.expected.LengthInBytes)
				}
				if bytesRead != tt.bytesRead {
					t.Errorf("Unmarshal() bytesRead = %v, want %v", bytesRead, tt.bytesRead)
				}
			}
		})
	}
}
