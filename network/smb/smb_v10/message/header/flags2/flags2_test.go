package flags2_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header/flags2"
)

func TestFlags2Values(t *testing.T) {
	testCases := []struct {
		name     string
		flag     uint16
		expected uint16
	}{
		{
			name:     "FLAGS2_LONG_NAMES_ALLOWED",
			flag:     flags2.FLAGS2_LONG_NAMES_ALLOWED,
			expected: flags2.FLAGS2_LONG_NAMES_ALLOWED,
		},
		{
			name:     "FLAGS2_LONG_NAMES_USED",
			flag:     flags2.FLAGS2_LONG_NAMES_USED,
			expected: flags2.FLAGS2_LONG_NAMES_USED,
		},
		{
			name:     "FLAGS2_EXTENDED_ATTRIBUTES",
			flag:     flags2.FLAGS2_EXTENDED_ATTRIBUTES,
			expected: flags2.FLAGS2_EXTENDED_ATTRIBUTES,
		},
		{
			name:     "FLAGS2_SECURITY_SIGNATURE",
			flag:     flags2.FLAGS2_SECURITY_SIGNATURE,
			expected: flags2.FLAGS2_SECURITY_SIGNATURE,
		},
		{
			name:     "FLAGS2_COMPRESSED",
			flag:     flags2.FLAGS2_COMPRESSED,
			expected: flags2.FLAGS2_COMPRESSED,
		},
		{
			name:     "FLAGS2_SECURITY_SIGNATURE_REQUIRED",
			flag:     flags2.FLAGS2_SECURITY_SIGNATURE_REQUIRED,
			expected: flags2.FLAGS2_SECURITY_SIGNATURE_REQUIRED,
		},
		{
			name:     "FLAGS2_REPARSE_PATH",
			flag:     flags2.FLAGS2_REPARSE_PATH,
			expected: flags2.FLAGS2_REPARSE_PATH,
		},
		{
			name:     "FLAGS2_EXTENDED_SECURITY",
			flag:     flags2.FLAGS2_EXTENDED_SECURITY,
			expected: flags2.FLAGS2_EXTENDED_SECURITY,
		},
		{
			name:     "FLAGS2_DFS",
			flag:     flags2.FLAGS2_DFS,
			expected: flags2.FLAGS2_DFS,
		},
		{
			name:     "FLAGS2_PAGING_IO",
			flag:     flags2.FLAGS2_PAGING_IO,
			expected: flags2.FLAGS2_PAGING_IO,
		},
		{
			name:     "FLAGS2_NT_STATUS_ERROR_CODES",
			flag:     flags2.FLAGS2_NT_STATUS_ERROR_CODES,
			expected: flags2.FLAGS2_NT_STATUS_ERROR_CODES,
		},
		{
			name:     "FLAGS2_UNICODE",
			flag:     flags2.FLAGS2_UNICODE,
			expected: flags2.FLAGS2_UNICODE,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.flag != tc.expected {
				t.Errorf("Expected %s to be 0x%04x, got 0x%04x", tc.name, tc.expected, tc.flag)
			}
		})
	}
}

func TestFlags2Combinations(t *testing.T) {
	// Test some common flag combinations
	testCases := []struct {
		name     string
		flags    uint16
		expected uint16
	}{
		{
			name:     "Unicode with long names",
			flags:    flags2.FLAGS2_UNICODE | flags2.FLAGS2_LONG_NAMES_USED,
			expected: 0x8080,
		},
		{
			name:     "NT Status with extended security",
			flags:    flags2.FLAGS2_NT_STATUS_ERROR_CODES | flags2.FLAGS2_EXTENDED_SECURITY,
			expected: 0x4800,
		},
		{
			name:     "Common client request flags",
			flags:    flags2.FLAGS2_UNICODE | flags2.FLAGS2_NT_STATUS_ERROR_CODES | flags2.FLAGS2_LONG_NAMES_USED | flags2.FLAGS2_SECURITY_SIGNATURE,
			expected: 0xc084,
		},
		{
			name: "All flags set",
			flags: flags2.FLAGS2_LONG_NAMES_ALLOWED | flags2.FLAGS2_LONG_NAMES_USED | flags2.FLAGS2_EXTENDED_ATTRIBUTES |
				flags2.FLAGS2_SECURITY_SIGNATURE | flags2.FLAGS2_COMPRESSED | flags2.FLAGS2_SECURITY_SIGNATURE_REQUIRED |
				flags2.FLAGS2_REPARSE_PATH | flags2.FLAGS2_EXTENDED_SECURITY | flags2.FLAGS2_DFS | flags2.FLAGS2_PAGING_IO |
				flags2.FLAGS2_NT_STATUS_ERROR_CODES | flags2.FLAGS2_UNICODE | flags2.FLAGS2_RESERVED_4 | flags2.FLAGS2_RESERVED_6 |
				flags2.FLAGS2_RESERVED_8 | flags2.FLAGS2_RESERVED_9,
			expected: 0xffff,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.flags != tc.expected {
				t.Errorf("Expected flags combination to be 0x%04x, got 0x%04x", tc.expected, tc.flags)
			}
		})
	}
}

func TestFlags2_String(t *testing.T) {
	testCases := []struct {
		name     string
		flags2   flags2.Flags2
		expected string
	}{
		{
			name:     "No flags set",
			flags2:   flags2.Flags2(0x00),
			expected: "NONE",
		},
		{
			name:     "Single flag - UNICODE",
			flags2:   flags2.Flags2(flags2.FLAGS2_UNICODE),
			expected: "UNICODE",
		},
		{
			name:     "Single flag - LONG_NAMES_USED",
			flags2:   flags2.Flags2(flags2.FLAGS2_LONG_NAMES_USED),
			expected: "LONG_NAMES_USED",
		},
		{
			name:     "Two flags - UNICODE and LONG_NAMES_USED",
			flags2:   flags2.Flags2(flags2.FLAGS2_UNICODE | flags2.FLAGS2_LONG_NAMES_USED),
			expected: "LONG_NAMES_USED|UNICODE",
		},
		{
			name:     "Multiple flags",
			flags2:   flags2.Flags2(flags2.FLAGS2_DFS | flags2.FLAGS2_EXTENDED_SECURITY | flags2.FLAGS2_NT_STATUS_ERROR_CODES),
			expected: "DFS|EXTENDED_SECURITY|NT_STATUS_ERROR_CODES",
		},
		{
			name:     "Reserved flags",
			flags2:   flags2.Flags2(flags2.FLAGS2_RESERVED_4 | flags2.FLAGS2_RESERVED_6 | flags2.FLAGS2_RESERVED_8 | flags2.FLAGS2_RESERVED_9),
			expected: "RESERVED_4|RESERVED_6|RESERVED_8|RESERVED_9",
		},
		{
			name: "Common client flags",
			flags2: flags2.Flags2(flags2.FLAGS2_UNICODE | flags2.FLAGS2_NT_STATUS_ERROR_CODES |
				flags2.FLAGS2_LONG_NAMES_USED | flags2.FLAGS2_EXTENDED_SECURITY),
			expected: "EXTENDED_SECURITY|LONG_NAMES_USED|NT_STATUS_ERROR_CODES|UNICODE",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.flags2.String()
			if result != tc.expected {
				t.Errorf("Expected String() to return %q, got %q", tc.expected, result)
			}
		})
	}
}
