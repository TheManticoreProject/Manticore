package flags_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header/flags"
)

func TestFlagsConstants(t *testing.T) {
	testCases := []struct {
		name     string
		flag     uint8
		expected uint8
	}{
		{
			name:     "FLAGS_LOCK_AND_READ_OK",
			flag:     flags.FLAGS_LOCK_AND_READ_OK,
			expected: 0x01,
		},
		{
			name:     "FLAGS_BUF_AVAIL",
			flag:     flags.FLAGS_BUF_AVAIL,
			expected: 0x02,
		},
		{
			name:     "FLAGS_RESERVED",
			flag:     flags.FLAGS_RESERVED,
			expected: 0x04,
		},
		{
			name:     "FLAGS_CASE_INSENSITIVE",
			flag:     flags.FLAGS_CASE_INSENSITIVE,
			expected: 0x08,
		},
		{
			name:     "FLAGS_CANONICALIZED_PATHS",
			flag:     flags.FLAGS_CANONICALIZED_PATHS,
			expected: 0x10,
		},
		{
			name:     "FLAGS_OPLOCK",
			flag:     flags.FLAGS_OPLOCK,
			expected: 0x20,
		},
		{
			name:     "FLAGS_OPBATCH",
			flag:     flags.FLAGS_OPBATCH,
			expected: 0x40,
		},
		{
			name:     "FLAGS_REPLY",
			flag:     flags.FLAGS_REPLY,
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
			flags:    flags.FLAGS_REPLY | flags.FLAGS_LOCK_AND_READ_OK,
			expected: 0x81,
		},
		{
			name:     "Oplock with canonicalized paths",
			flags:    flags.FLAGS_OPLOCK | flags.FLAGS_CANONICALIZED_PATHS,
			expected: 0x30,
		},
		{
			name:     "All flags set",
			flags:    flags.FLAGS_LOCK_AND_READ_OK | flags.FLAGS_BUF_AVAIL | flags.FLAGS_RESERVED | flags.FLAGS_CASE_INSENSITIVE | flags.FLAGS_CANONICALIZED_PATHS | flags.FLAGS_OPLOCK | flags.FLAGS_OPBATCH | flags.FLAGS_REPLY,
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

func TestFlags_String(t *testing.T) {
	testCases := []struct {
		name     string
		flags    flags.Flags
		expected string
	}{
		{
			name:     "No flags set",
			flags:    flags.Flags(0x00),
			expected: "NONE",
		},
		{
			name:     "Single flag - REPLY",
			flags:    flags.Flags(flags.FLAGS_REPLY),
			expected: "REPLY",
		},
		{
			name:     "Single flag - LOCK_AND_READ_OK",
			flags:    flags.Flags(flags.FLAGS_LOCK_AND_READ_OK),
			expected: "LOCK_AND_READ_OK",
		},
		{
			name:     "Two flags - REPLY and LOCK_AND_READ_OK",
			flags:    flags.Flags(flags.FLAGS_REPLY | flags.FLAGS_LOCK_AND_READ_OK),
			expected: "LOCK_AND_READ_OK|REPLY",
		},
		{
			name:     "Multiple flags",
			flags:    flags.Flags(flags.FLAGS_OPLOCK | flags.FLAGS_CANONICALIZED_PATHS | flags.FLAGS_CASE_INSENSITIVE),
			expected: "CANONICALIZED_PATHS|CASE_INSENSITIVE|OPLOCK",
		},
		{
			name:     "All flags set",
			flags:    flags.Flags(flags.FLAGS_LOCK_AND_READ_OK | flags.FLAGS_BUF_AVAIL | flags.FLAGS_RESERVED | flags.FLAGS_CASE_INSENSITIVE | flags.FLAGS_CANONICALIZED_PATHS | flags.FLAGS_OPLOCK | flags.FLAGS_OPBATCH | flags.FLAGS_REPLY),
			expected: "BUF_AVAIL|CANONICALIZED_PATHS|CASE_INSENSITIVE|LOCK_AND_READ_OK|OPBATCH|OPLOCK|REPLY|RESERVED",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.flags.String()
			if result != tc.expected {
				t.Errorf("Expected String() to return %q, got %q", tc.expected, result)
			}
		})
	}
}
