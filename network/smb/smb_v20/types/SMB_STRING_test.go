package types

import (
	"fmt"
	"testing"
)

func TestSMB_STRING_Marshal(t *testing.T) {
	testCases := []struct {
		name     string
		input    *SMB_STRING
		expected []byte
	}{
		{
			name: "Empty string with format 0x04",
			input: &SMB_STRING{
				BufferFormat: 0x04,
				Length:       0,
				Buffer:       []UCHAR{},
			},
			expected: []byte{0x04, 0x00},
		},
		{
			name: "Simple string with format 0x04",
			input: &SMB_STRING{
				BufferFormat: 0x04,
				Length:       11,
				Buffer:       []UCHAR("hello world"),
			},
			expected: []byte{0x04, 'h', 'e', 'l', 'l', 'o', ' ', 'w', 'o', 'r', 'l', 'd', 0x00},
		},
		{
			name: "Special characters with format 0x04",
			input: &SMB_STRING{
				BufferFormat: 0x04,
				Length:       5,
				Buffer:       []UCHAR("a@b#c"),
			},
			expected: []byte{0x04, 'a', '@', 'b', '#', 'c', 0x00},
		},
		{
			name: "String with format 0x01",
			input: &SMB_STRING{
				BufferFormat: 0x01,
				Length:       5,
				Buffer:       []UCHAR("hello"),
			},
			expected: []byte{0x01, 0x06, 0x00, 'h', 'e', 'l', 'l', 'o', 0x00},
		},
		{
			name: "String with format 0x02",
			input: &SMB_STRING{
				BufferFormat: 0x02,
				Length:       11,
				Buffer:       []UCHAR("NT LM 0.12"),
			},
			expected: []byte{0x02, 'N', 'T', ' ', 'L', 'M', ' ', '0', '.', '1', '2', 0x00},
		},
		{
			name: "String with format 0x03",
			input: &SMB_STRING{
				BufferFormat: 0x03,
				Length:       8,
				Buffer:       []UCHAR("username"),
			},
			expected: []byte{0x03, 0x09, 0x00, 'u', 's', 'e', 'r', 'n', 'a', 'm', 'e', 0x00},
		},
		{
			name: "String with format 0x05",
			input: &SMB_STRING{
				BufferFormat: 0x05,
				Length:       4,
				Buffer:       []UCHAR("path"),
			},
			expected: []byte{0x05, 0x05, 0x00, 'p', 'a', 't', 'h', 0x00},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := tc.input.Marshal()
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(result) != len(tc.expected) {
				fmt.Printf(" | Expected : %x\n", tc.expected)
				fmt.Printf(" | Result   : %x\n", result)
				t.Fatalf("Expected result length %d, got %d", len(tc.expected), len(result))
			}

			for i, b := range tc.expected {
				if result[i] != b {
					fmt.Printf(" | Expected : %x\n", tc.expected)
					fmt.Printf(" | Result   : %x\n", result)
					t.Errorf("Mismatch at index %d: expected %d, got %d", i, b, result[i])
				}
			}
		})
	}
}

func TestSMB_STRING_Unmarshal(t *testing.T) {
	testCases := []struct {
		name           string
		input          []byte
		expectedFormat UCHAR
		expectedLength USHORT
		expectedBuffer []UCHAR
		expectError    bool
	}{
		{
			name:           "Format 0x04 null-terminated string",
			input:          []byte{0x04, 'h', 'e', 'l', 'l', 'o', 0x00},
			expectedFormat: 0x04,
			expectedLength: 5,
			expectedBuffer: []UCHAR("hello"),
			expectError:    false,
		},
		{
			name:           "Format 0x01 variable block with length",
			input:          []byte{0x01, 0x05, 0x00, 'h', 'e', 'l', 'l', 'o'},
			expectedFormat: 0x01,
			expectedLength: 5,
			expectedBuffer: []UCHAR("hello"),
			expectError:    false,
		},
		{
			name:           "Format 0x02 null-terminated dialect string",
			input:          []byte{0x02, 'N', 'T', ' ', 'L', 'M', ' ', '0', '.', '1', '2', 0x00},
			expectedFormat: 0x02,
			expectedLength: 10,
			expectedBuffer: []UCHAR("NT LM 0.12"),
			expectError:    false,
		},
		{
			name:           "Format 0x03 variable block with length",
			input:          []byte{0x03, 0x04, 0x00, 't', 'e', 's', 't'},
			expectedFormat: 0x03,
			expectedLength: 4,
			expectedBuffer: []UCHAR("test"),
			expectError:    false,
		},
		{
			name:           "Format 0x05 null-terminated dialect string",
			input:          []byte{0x05, 'P', 'C', ' ', 'N', 'E', 'T', 'W', 'O', 'R', 'K', ' ', 'P', 'R', 'O', 'G', 'R', 'A', 'M', 0x00},
			expectedFormat: 0x05,
			expectedLength: 18,
			expectedBuffer: []UCHAR("PC NETWORK PROGRAM"),
			expectError:    false,
		},
		{
			name:        "Invalid buffer format",
			input:       []byte{0x06, 'i', 'n', 'v', 'a', 'l', 'i', 'd', 0x00},
			expectError: true,
		},
		{
			name:        "Buffer too short",
			input:       []byte{0x04},
			expectError: true,
		},
		{
			name:        "Format 0x01 with insufficient data",
			input:       []byte{0x01, 0x10, 0x00, 't', 'o', 'o', ' ', 's', 'h', 'o', 'r', 't'},
			expectError: true,
		},
		{
			name:        "Format 0x04 with no null terminator",
			input:       []byte{0x04, 'n', 'o', 't', 'e', 'r', 'm', 'i', 'n', 'a', 't', 'e', 'd'},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			smbString := &SMB_STRING{}
			_, err := smbString.Unmarshal(tc.input)

			if tc.expectError {
				if err == nil {
					t.Fatalf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if smbString.BufferFormat != tc.expectedFormat {
				t.Errorf("Expected BufferFormat %d, got %d", tc.expectedFormat, smbString.BufferFormat)
			}

			if smbString.Length != tc.expectedLength {
				t.Errorf("Expected Length %d, got %d", tc.expectedLength, smbString.Length)
			}

			if len(smbString.Buffer) != len(tc.expectedBuffer) {
				t.Errorf("Expected Buffer length %d, got %d", len(tc.expectedBuffer), len(smbString.Buffer))
			} else {
				for i, b := range tc.expectedBuffer {
					if smbString.Buffer[i] != b {
						t.Errorf("Buffer mismatch at index %d: expected %d, got %d", i, b, smbString.Buffer[i])
					}
				}
			}
		})
	}
}
