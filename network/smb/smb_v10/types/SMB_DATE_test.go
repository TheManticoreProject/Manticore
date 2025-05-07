package types_test

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

func TestSMB_DATE_NewSMB_DATE(t *testing.T) {
	date := types.NewSMB_DATE()

	if date.Year != 0 {
		t.Errorf("Expected Year to be 0, got %d", date.Year)
	}

	if date.Month != 0 {
		t.Errorf("Expected Month to be 0, got %d", date.Month)
	}

	if date.Day != 0 {
		t.Errorf("Expected Day to be 0, got %d", date.Day)
	}
}

func TestSMB_DATE_Marshal(t *testing.T) {
	tests := []struct {
		name     string
		input    types.SMB_DATE
		expected []byte
	}{
		{
			name: "Basic Date",
			input: types.SMB_DATE{
				Year:  2023,
				Month: 5,
				Day:   15,
			},
			expected: []byte{
				0xAD, 0x54, // (2023-1980)<<9 | 5<<5 | 15 = 0x54AD
			},
		},
		{
			name: "Minimum Date",
			input: types.SMB_DATE{
				Year:  1980,
				Month: 1,
				Day:   1,
			},
			expected: []byte{
				0x21, 0x00, // (1980-1980)<<9 | 1<<5 | 1 = 0x0021
			},
		},
		{
			name: "Maximum Values",
			input: types.SMB_DATE{
				Year:  2107, // Max year (127 + 1980)
				Month: 12,
				Day:   31,
			},
			expected: []byte{
				0x9F, 0xFE, // (2107-1980)<<9 | 12<<5 | 31 = 0xFE9F
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

func TestSMB_DATE_Unmarshal(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected types.SMB_DATE
	}{
		{
			name:  "Basic Date",
			input: []byte{0xAD, 0x54}, // (2023-1980)<<9 | 5<<5 | 15 = 0x54AD
			expected: types.SMB_DATE{
				Year:  2023,
				Month: 5,
				Day:   15,
			},
		},
		{
			name:  "Minimum Date",
			input: []byte{0x21, 0x00}, // (1980-1980)<<9 | 1<<5 | 1 = 0x0021
			expected: types.SMB_DATE{
				Year:  1980,
				Month: 1,
				Day:   1,
			},
		},
		{
			name:  "Maximum Values",
			input: []byte{0x9F, 0xFE}, // (2107-1980)<<9 | 12<<5 | 31 = 0xFE9F
			expected: types.SMB_DATE{
				Year:  2107,
				Month: 12,
				Day:   31,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			date := types.NewSMB_DATE()
			bytesRead, err := date.Unmarshal(tt.input)
			if err != nil {
				t.Errorf("Unmarshal() error = %v", err)
				return
			}
			if bytesRead != len(tt.input) {
				t.Errorf("Unmarshal() bytesRead = %v, want %v", bytesRead, len(tt.input))
			}
			if date.Year != tt.expected.Year {
				t.Errorf("Unmarshal() Year = %v, want %v", date.Year, tt.expected.Year)
			}
			if date.Month != tt.expected.Month {
				t.Errorf("Unmarshal() Month = %v, want %v", date.Month, tt.expected.Month)
			}
			if date.Day != tt.expected.Day {
				t.Errorf("Unmarshal() Day = %v, want %v", date.Day, tt.expected.Day)
			}
		})
	}
}
