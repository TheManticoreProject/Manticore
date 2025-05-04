package utils_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/utils"
)

func TestPadStringLeft(t *testing.T) {
	tests := []struct {
		input    string
		padChar  string
		length   int
		expected string
	}{
		{"hello", "*", 8, "***hello"},
		{"world", "-", 10, "-----world"},
		{"test", " ", 6, "  test"},
		{"", "#", 5, "#####"},
		{"short", "0", 5, "short"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := utils.PadStringLeft(tt.input, tt.padChar, tt.length)
			if result != tt.expected {
				t.Errorf("PadStringLeft(%q, %q, %d) = %q; want %q", tt.input, tt.padChar, tt.length, result, tt.expected)
			}
		})
	}
}

func TestSizeInBytes(t *testing.T) {
	tests := []struct {
		size     uint64
		expected string
	}{
		{512, "512 bytes"},
		{1024, "1.00 KiB"},
		{1048576, "1.00 MiB"},
		{1073741824, "1.00 GiB"},
		{1099511627776, "1.00 TiB"},
		{1125899906842624, "1.00 PiB"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := utils.SizeInBytes(tt.size)
			if result != tt.expected {
				t.Errorf("SizeInBytes(%d) = %q; want %q", tt.size, result, tt.expected)
			}
		})
	}
}
