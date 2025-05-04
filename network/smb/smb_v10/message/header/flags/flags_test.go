package flags

import (
	"testing"
)

func TestFlagsConstants(t *testing.T) {
	testCases := []struct {
		name     string
		flag     uint8
		expected uint8
	}{
		{
			name:     "FLAGS_LOCK_AND_READ_OK",
			flag:     FLAGS_LOCK_AND_READ_OK,
			expected: 0x01,
		},
		{
			name:     "FLAGS_BUF_AVAIL",
			flag:     FLAGS_BUF_AVAIL,
			expected: 0x02,
		},
		{
			name:     "FLAGS_RESERVED",
			flag:     FLAGS_RESERVED,
			expected: 0x04,
		},
		{
			name:     "FLAGS_CASE_INSENSITIVE",
			flag:     FLAGS_CASE_INSENSITIVE,
			expected: 0x08,
		},
		{
			name:     "FLAGS_CANONICALIZED_PATHS",
			flag:     FLAGS_CANONICALIZED_PATHS,
			expected: 0x10,
		},
		{
			name:     "FLAGS_OPLOCK",
			flag:     FLAGS_OPLOCK,
			expected: 0x20,
		},
		{
			name:     "FLAGS_OPBATCH",
			flag:     FLAGS_OPBATCH,
			expected: 0x40,
		},
		{
			name:     "FLAGS_REPLY",
			flag:     FLAGS_REPLY,
			expected: 0x80,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.flag != tc.expected {
				t.Errorf("Expected %s to be 0x%02x, got 0x%02x", tc.name, tc.expected, tc.flag)
			}
		})
	}
}

func TestFlagsCombinations(t *testing.T) {
	// Test some common flag combinations
	testCases := []struct {
		name     string
		flags    uint8
		expected uint8
	}{
		{
			name:     "Server reply with lock and read",
			flags:    FLAGS_REPLY | FLAGS_LOCK_AND_READ_OK,
			expected: 0x81,
		},
		{
			name:     "Oplock with canonicalized paths",
			flags:    FLAGS_OPLOCK | FLAGS_CANONICALIZED_PATHS,
			expected: 0x30,
		},
		{
			name:     "All flags set",
			flags:    FLAGS_LOCK_AND_READ_OK | FLAGS_BUF_AVAIL | FLAGS_RESERVED | FLAGS_CASE_INSENSITIVE | FLAGS_CANONICALIZED_PATHS | FLAGS_OPLOCK | FLAGS_OPBATCH | FLAGS_REPLY,
			expected: 0xFF,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.flags != tc.expected {
				t.Errorf("Expected flags combination to be 0x%02x, got 0x%02x", tc.expected, tc.flags)
			}
		})
	}
}
