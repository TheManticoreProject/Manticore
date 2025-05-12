package utils_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/utils"
)

func TestGetNullTerminatedString(t *testing.T) {
	tests := []struct {
		name           string
		input          []byte
		expectedString string
		expectedOffset int
	}{
		{
			name:           "Empty string",
			input:          []byte{0},
			expectedString: "",
			expectedOffset: 1,
		},
		{
			name:           "Simple string",
			input:          []byte{'H', 'e', 'l', 'l', 'o', 0},
			expectedString: "Hello",
			expectedOffset: 6,
		},
		{
			name:           "String with data after null terminator",
			input:          []byte{'H', 'e', 'l', 'l', 'o', 0, 'W', 'o', 'r', 'l', 'd'},
			expectedString: "Hello",
			expectedOffset: 6,
		},
		{
			name:           "String with special characters",
			input:          []byte{'T', 'e', 's', 't', '!', '@', '#', '$', '%', 0},
			expectedString: "Test!@#$%",
			expectedOffset: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			str, offset := utils.GetNullTerminatedString(tt.input)
			if str != tt.expectedString {
				t.Errorf("GetNullTerminatedString() got string = %v, want %v", str, tt.expectedString)
			}
			if offset != tt.expectedOffset {
				t.Errorf("GetNullTerminatedString() got offset = %v, want %v", offset, tt.expectedOffset)
			}
		})
	}
}

func TestGetNullTerminatedUnicodeString(t *testing.T) {
	tests := []struct {
		name           string
		input          []byte
		expectedString string
		expectedOffset int
	}{
		{
			name:           "Empty unicode string",
			input:          []byte{0, 0},
			expectedString: "",
			expectedOffset: 2,
		},
		{
			name:           "Simple unicode string",
			input:          []byte{'H', 0, 'e', 0, 'l', 0, 'l', 0, 'o', 0, 0, 0},
			expectedString: "H\x00e\x00l\x00l\x00o\x00",
			expectedOffset: 12,
		},
		{
			name:           "Unicode string with data after null terminator",
			input:          []byte{'H', 0, 'e', 0, 'l', 0, 'l', 0, 'o', 0, 0, 0, 'W', 0, 'o', 0, 'r', 0, 'l', 0, 'd', 0},
			expectedString: "H\x00e\x00l\x00l\x00o\x00",
			expectedOffset: 12,
		},
		{
			name:           "Unicode string with special characters",
			input:          []byte{'T', 0, 'e', 0, 's', 0, 't', 0, '!', 0, '@', 0, '#', 0, '$', 0, '%', 0, 0, 0},
			expectedString: "T\x00e\x00s\x00t\x00!\x00@\x00#\x00$\x00%\x00",
			expectedOffset: 20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			str, offset := utils.GetNullTerminatedUnicodeString(tt.input)
			if str != tt.expectedString {
				t.Errorf("GetNullTerminatedUnicodeString() got string = %v, want %v", str, tt.expectedString)
			}
			if offset != tt.expectedOffset {
				t.Errorf("GetNullTerminatedUnicodeString() got offset = %v, want %v", offset, tt.expectedOffset)
			}
		})
	}
}
