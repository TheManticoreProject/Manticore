package types_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
	"github.com/TheManticoreProject/Manticore/utils/encoding/utf16"
)

func TestSMB_STRING_Marshal(t *testing.T) {
	testCases := []struct {
		name     string
		input    *types.SMB_STRING
		expected []byte
	}{
		{
			name: "Empty string with format 0x01 (VARIABLE_BLOCK_16BIT)",
			input: &types.SMB_STRING{
				BufferFormat: types.SMB_STRING_BUFFER_FORMAT_VARIABLE_BLOCK_16BIT,
				Length:       0,
				Buffer:       []types.UCHAR{},
			},
			expected: []byte{
				0x01,       // Format
				0x00, 0x00, // Length
			},
		},
		{
			name: "Empty string with format 0x02 (NULL_TERMINATED_OEM_STRING)",
			input: &types.SMB_STRING{
				BufferFormat: types.SMB_STRING_BUFFER_FORMAT_NULL_TERMINATED_OEM_STRING,
				Length:       0,
				Buffer:       []types.UCHAR{},
			},
			expected: []byte{
				0x02, // Format
				0x00, // Null terminator
			},
		},
		{
			name: "Empty string with format 0x03 (NULL_TERMINATED_OEM_STRING_16BIT)",
			input: &types.SMB_STRING{
				BufferFormat: types.SMB_STRING_BUFFER_FORMAT_NULL_TERMINATED_OEM_STRING_16BIT,
				Length:       0,
				Buffer:       []types.UCHAR{},
			},
			expected: []byte{
				0x03,       // Format
				0x00, 0x00, // Length
				0x00, // Null terminator
			},
		},
		{
			name: "Empty string with format 0x04 (NULL_TERMINATED_ASCII_STRING)",
			input: &types.SMB_STRING{
				BufferFormat: types.SMB_STRING_BUFFER_FORMAT_NULL_TERMINATED_ASCII_STRING,
				Length:       0,
				Buffer:       []types.UCHAR{},
			},
			expected: []byte{
				0x04, // Format
				0x00, // Null terminator
			},
		},
		{
			name: "Empty string with format 0x05 (VARIABLE_BLOCK)",
			input: &types.SMB_STRING{
				BufferFormat: types.SMB_STRING_BUFFER_FORMAT_VARIABLE_BLOCK,
				Length:       0,
				Buffer:       []types.UCHAR{},
			},
			expected: []byte{
				0x05,       // Format
				0x00, 0x00, // Length
			},
		},
		{
			name: "String with format 0x01 (VARIABLE_BLOCK_16BIT)",
			input: &types.SMB_STRING{
				BufferFormat: types.SMB_STRING_BUFFER_FORMAT_VARIABLE_BLOCK_16BIT,
				Length:       10,
				Buffer:       []types.UCHAR(utf16.EncodeUTF16LE("hello")),
			},
			expected: []byte{
				0x01,       // Format
				0x0A, 0x00, // Length
				'h', 0x00, 'e', 0x00, 'l', 0x00, 'l', 0x00, 'o', 0x00, // Buffer
			},
		},
		{
			name: "String with format 0x02 (NULL_TERMINATED_OEM_STRING)",
			input: &types.SMB_STRING{
				BufferFormat: types.SMB_STRING_BUFFER_FORMAT_NULL_TERMINATED_OEM_STRING,
				Length:       11,
				Buffer:       []types.UCHAR("NT LM 0.12"),
			},
			expected: []byte{
				0x02,                                             // Format
				'N', 'T', ' ', 'L', 'M', ' ', '0', '.', '1', '2', // Buffer
				0x00, // Null terminator
			},
		},
		{
			name: "String with format 0x03 (NULL_TERMINATED_OEM_STRING_16BIT)",
			input: &types.SMB_STRING{
				BufferFormat: types.SMB_STRING_BUFFER_FORMAT_NULL_TERMINATED_OEM_STRING_16BIT,
				Length:       16,
				Buffer:       []types.UCHAR(utf16.EncodeUTF16LE("username")),
			},
			expected: []byte{
				0x03,       // Format
				0x10, 0x00, // Length
				'u', 0x00, 's', 0x00, 'e', 0x00, 'r', 0x00, 'n', 0x00, 'a', 0x00, 'm', 0x00, 'e', 0x00, // Buffer
				0x00, // Null terminator
			},
		},
		{
			name: "Simple string with format 0x04 (NULL_TERMINATED_ASCII_STRING)",
			input: &types.SMB_STRING{
				BufferFormat: types.SMB_STRING_BUFFER_FORMAT_NULL_TERMINATED_ASCII_STRING,
				Length:       11,
				Buffer:       []types.UCHAR("hello world"),
			},
			expected: []byte{
				0x04,                                                  // Format
				'h', 'e', 'l', 'l', 'o', ' ', 'w', 'o', 'r', 'l', 'd', // Buffer
				0x00, // Null terminator
			},
		},
		{
			name: "String with format 0x05 (VARIABLE_BLOCK)",
			input: &types.SMB_STRING{
				BufferFormat: types.SMB_STRING_BUFFER_FORMAT_VARIABLE_BLOCK,
				Length:       4,
				Buffer:       []types.UCHAR("path"),
			},
			expected: []byte{
				0x05,       // Format
				0x04, 0x00, // Length
				'p', 'a', 't', 'h', // Buffer
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := tc.input.Marshal()
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(result) != len(tc.expected) {
				t.Fatalf("Expected result length %d, got %d", len(tc.expected), len(result))
			}

			if !bytes.Equal(result, tc.expected) {
				t.Errorf("Buffer mismatch: expected %x, got %x", tc.expected, result)
			}
		})
	}
}

func TestSMB_STRING_Unmarshal(t *testing.T) {
	testCases := []struct {
		name           string
		input          []byte
		expectedFormat types.UCHAR
		expectedLength types.USHORT
		expectedBuffer []types.UCHAR
		expectError    bool
	}{
		{
			name: "Format 0x01 (VARIABLE_BLOCK_16BIT)",
			input: []byte{
				types.SMB_STRING_BUFFER_FORMAT_VARIABLE_BLOCK_16BIT, // Format
				0x05, 0x00, // Length
				'h', 'e', 'l', 'l', 'o', // Buffer
			},
			expectedFormat: 0x01,
			expectedLength: 5,
			expectedBuffer: []types.UCHAR("hello"),
			expectError:    false,
		},
		{
			name: "Format 0x02 (NULL_TERMINATED_OEM_STRING)",
			input: []byte{
				types.SMB_STRING_BUFFER_FORMAT_NULL_TERMINATED_OEM_STRING, // Format
				'N', 'T', ' ', 'L', 'M', ' ', '0', '.', '1', '2', // Buffer
				0x00, // Null terminator
			},
			expectedFormat: 0x02,
			expectedLength: 10,
			expectedBuffer: []types.UCHAR("NT LM 0.12"),
			expectError:    false,
		},
		{
			name: "Format 0x03 (NULL_TERMINATED_OEM_STRING_16BIT)",
			input: []byte{
				types.SMB_STRING_BUFFER_FORMAT_NULL_TERMINATED_OEM_STRING_16BIT, // Format
				0x08, 0x00, // Length
				't', 0x00, 'e', 0x00, 's', 0x00, 't', 0x00, // Buffer
				0x00, // Null terminator
			},
			expectedFormat: 0x03,
			expectedLength: 8,
			expectedBuffer: []types.UCHAR(utf16.EncodeUTF16LE("test")),
			expectError:    false,
		},
		{
			name: "Format 0x04 (NULL_TERMINATED_ASCII_STRING)",
			input: []byte{
				types.SMB_STRING_BUFFER_FORMAT_NULL_TERMINATED_ASCII_STRING, // Format
				'h', 'e', 'l', 'l', 'o', ' ', 'w', 'o', 'r', 'l', 'd', // Buffer
				0x00, // Null terminator
			},
			expectedFormat: 0x04,
			expectedLength: 11,
			expectedBuffer: []types.UCHAR("hello world"),
			expectError:    false,
		},
		{
			name: "Format 0x05 (VARIABLE_BLOCK)",
			input: []byte{
				types.SMB_STRING_BUFFER_FORMAT_VARIABLE_BLOCK, // Format
				0x12, 0x00, // Length
				'P', 'C', ' ', 'N', 'E', 'T', 'W', 'O', 'R', 'K', ' ', 'P', 'R', 'O', 'G', 'R', 'A', 'M', // Buffer
				0x00, // Null terminator
			},
			expectedFormat: 0x05,
			expectedLength: 18,
			expectedBuffer: []types.UCHAR("PC NETWORK PROGRAM"),
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
			name:        "Format 0x01 (VARIABLE_BLOCK_16BIT) with insufficient data",
			input:       []byte{0x01, 0x10, 0x00, 't', 'o', 'o', ' ', 's', 'h', 'o', 'r', 't'},
			expectError: true,
		},
		{
			name:        "Format 0x04 (NULL_TERMINATED_ASCII_STRING) with no null terminator",
			input:       []byte{0x04, 'n', 'o', 't', 'e', 'r', 'm', 'i', 'n', 'a', 't', 'e', 'd'},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			smbString := &types.SMB_STRING{}
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
			}

			if !bytes.Equal(smbString.Buffer, tc.expectedBuffer) {
				t.Errorf("Expected Buffer %x, got %x", tc.expectedBuffer, smbString.Buffer)
			}
		})
	}
}

func TestMarshalUnmarshal(t *testing.T) {
	testCases := []struct {
		name         string
		bufferFormat types.UCHAR
		inputString  []byte
		expectError  bool
	}{
		{
			name:         "Format 0x01 (VARIABLE_BLOCK_16BIT)",
			bufferFormat: types.SMB_STRING_BUFFER_FORMAT_VARIABLE_BLOCK_16BIT,
			inputString:  utf16.EncodeUTF16LE("test string"),
			expectError:  false,
		},
		{
			name:         "Format 0x02 (NULL_TERMINATED_OEM_STRING)",
			bufferFormat: types.SMB_STRING_BUFFER_FORMAT_NULL_TERMINATED_OEM_STRING,
			inputString:  []byte("dialect string"),
			expectError:  false,
		},
		{
			name:         "Format 0x03 (NULL_TERMINATED_OEM_STRING_16BIT)",
			bufferFormat: types.SMB_STRING_BUFFER_FORMAT_NULL_TERMINATED_OEM_STRING_16BIT,
			inputString:  utf16.EncodeUTF16LE("oem string with length"),
			expectError:  false,
		},
		{
			name:         "Format 0x04 (NULL_TERMINATED_ASCII_STRING)",
			bufferFormat: types.SMB_STRING_BUFFER_FORMAT_NULL_TERMINATED_ASCII_STRING,
			inputString:  []byte("ascii string"),
			expectError:  false,
		},
		{
			name:         "Format 0x05 (VARIABLE_BLOCK)",
			bufferFormat: types.SMB_STRING_BUFFER_FORMAT_VARIABLE_BLOCK,
			inputString:  []byte("variable block"),
			expectError:  false,
		},
		{
			name:         "Invalid Format",
			bufferFormat: 0xFF,
			inputString:  []byte("invalid format"),
			expectError:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create and setup the original SMB_STRING
			original := &types.SMB_STRING{}
			original.SetBufferFormat(tc.bufferFormat)
			original.SetString(string(tc.inputString))

			// Marshal the SMB_STRING
			marshalled, err := original.Marshal()
			if tc.expectError {
				if err == nil {
					t.Fatalf("Expected marshal error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected marshal error: %v", err)
			}

			// Unmarshal back into a new SMB_STRING
			unmarshalled := &types.SMB_STRING{}
			bytesRead, err := unmarshalled.Unmarshal(marshalled)
			if err != nil {
				t.Fatalf("Unexpected unmarshal error: %v", err)
			}

			// Verify the unmarshalled data matches the original
			if unmarshalled.BufferFormat != original.BufferFormat {
				t.Errorf("BufferFormat mismatch: expected %d, got %d", original.BufferFormat, unmarshalled.BufferFormat)
			}

			if unmarshalled.Length != original.Length {
				t.Errorf("Length mismatch: expected %d, got %d", original.Length, unmarshalled.Length)
			}

			if len(unmarshalled.Buffer) != len(original.Buffer) {
				t.Errorf("Buffer length mismatch: expected %d, got %d", len(original.Buffer), len(unmarshalled.Buffer))
			}
			if !bytes.Equal(unmarshalled.Buffer, original.Buffer) {
				t.Errorf("Buffer mismatch: expected %x, got %x", original.Buffer, unmarshalled.Buffer)
			}

			// Verify all bytes were consumed
			if bytesRead != len(marshalled) {
				unmarshalledBytes, err := unmarshalled.Marshal()
				if err != nil {
					t.Fatalf("Unexpected marshal error: %v", err)
				}
				fmt.Printf("marshalled   : %x (%d)\n", marshalled, len(marshalled))
				fmt.Printf("unmarshalled : %x (%d)\n", unmarshalledBytes, len(unmarshalledBytes))

				t.Errorf("Not all bytes consumed: expected %d, got %d", len(marshalled), bytesRead)
			}
		})
	}
}
