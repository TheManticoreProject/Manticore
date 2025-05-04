package flags

import (
	"testing"
)

func TestFlags2Values(t *testing.T) {
	testCases := []struct {
		name     string
		flag     uint16
		expected uint16
	}{
		{
			name:     "FLAGS2_LONG_NAMES_USED",
			flag:     FLAGS2_LONG_NAMES_USED,
			expected: FLAGS2_LONG_NAMES_USED,
		},
		{
			name:     "FLAGS2_EXTENDED_ATTRIBUTES",
			flag:     FLAGS2_EXTENDED_ATTRIBUTES,
			expected: FLAGS2_EXTENDED_ATTRIBUTES,
		},
		{
			name:     "FLAGS2_SMB_SECURITY_SIGNATURE",
			flag:     FLAGS2_SMB_SECURITY_SIGNATURE,
			expected: FLAGS2_SMB_SECURITY_SIGNATURE,
		},
		{
			name:     "FLAGS2_EXTENDED_SECURITY",
			flag:     FLAGS2_EXTENDED_SECURITY,
			expected: FLAGS2_EXTENDED_SECURITY,
		},
		{
			name:     "FLAGS2_DFS",
			flag:     FLAGS2_DFS,
			expected: FLAGS2_DFS,
		},
		{
			name:     "FLAGS2_PAGING_IO",
			flag:     FLAGS2_PAGING_IO,
			expected: FLAGS2_PAGING_IO,
		},
		{
			name:     "FLAGS2_UNICODE",
			flag:     FLAGS2_UNICODE,
			expected: FLAGS2_UNICODE,
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
			flags:    FLAGS2_UNICODE | FLAGS2_LONG_NAMES_USED,
			expected: 0x8080,
		},
		{
			name:     "NT Status with extended security",
			flags:    FLAGS2_NT_STATUS_ERROR_CODES | FLAGS2_EXTENDED_SECURITY,
			expected: 0x4800,
		},
		{
			name:     "Common client request flags",
			flags:    FLAGS2_UNICODE | FLAGS2_NT_STATUS_ERROR_CODES | FLAGS2_LONG_NAMES_USED | FLAGS2_SMB_SECURITY_SIGNATURE,
			expected: 0xc084,
		},
		{
			name: "All flags set",
			flags: FLAGS2_LONG_NAMES_USED | FLAGS2_EXTENDED_ATTRIBUTES | FLAGS2_SMB_SECURITY_SIGNATURE |
				FLAGS2_EXTENDED_SECURITY | FLAGS2_DFS | FLAGS2_PAGING_IO | FLAGS2_NT_STATUS_ERROR_CODES | FLAGS2_UNICODE,
			expected: 0xf886,
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
