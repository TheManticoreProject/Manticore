package subcommands_test

import (
	"testing"

	smb_v10 "github.com/TheManticoreProject/Manticore/network/smb/smb_v10/subcommands"
)

func TestTransaction2SubcommandString(t *testing.T) {
	tests := []struct {
		name     string
		subCmd   smb_v10.Transaction2Subcommand
		expected string
	}{
		{
			name:     "OPEN2",
			subCmd:   smb_v10.TRANS2_OPEN2,
			expected: "OPEN2",
		},
		{
			name:     "FIND_FIRST2",
			subCmd:   smb_v10.TRANS2_FIND_FIRST2,
			expected: "FIND_FIRST2",
		},
		{
			name:     "FIND_NEXT2",
			subCmd:   smb_v10.TRANS2_FIND_NEXT2,
			expected: "FIND_NEXT2",
		},
		{
			name:     "QUERY_FS_INFORMATION",
			subCmd:   smb_v10.TRANS2_QUERY_FS_INFORMATION,
			expected: "QUERY_FS_INFORMATION",
		},
		{
			name:     "SET_FS_INFORMATION",
			subCmd:   smb_v10.TRANS2_SET_FS_INFORMATION,
			expected: "SET_FS_INFORMATION",
		},
		{
			name:     "QUERY_PATH_INFORMATION",
			subCmd:   smb_v10.TRANS2_QUERY_PATH_INFORMATION,
			expected: "QUERY_PATH_INFORMATION",
		},
		{
			name:     "SET_PATH_INFORMATION",
			subCmd:   smb_v10.TRANS2_SET_PATH_INFORMATION,
			expected: "SET_PATH_INFORMATION",
		},
		{
			name:     "QUERY_FILE_INFORMATION",
			subCmd:   smb_v10.TRANS2_QUERY_FILE_INFORMATION,
			expected: "QUERY_FILE_INFORMATION",
		},
		{
			name:     "SET_FILE_INFORMATION",
			subCmd:   smb_v10.TRANS2_SET_FILE_INFORMATION,
			expected: "SET_FILE_INFORMATION",
		},
		{
			name:     "FSCTL",
			subCmd:   smb_v10.TRANS2_FSCTL,
			expected: "FSCTL",
		},
		{
			name:     "IOCTL2",
			subCmd:   smb_v10.TRANS2_IOCTL2,
			expected: "IOCTL2",
		},
		{
			name:     "FIND_NOTIFY_FIRST",
			subCmd:   smb_v10.TRANS2_FIND_NOTIFY_FIRST,
			expected: "FIND_NOTIFY_FIRST",
		},
		{
			name:     "FIND_NOTIFY_NEXT",
			subCmd:   smb_v10.TRANS2_FIND_NOTIFY_NEXT,
			expected: "FIND_NOTIFY_NEXT",
		},
		{
			name:     "CREATE_DIRECTORY",
			subCmd:   smb_v10.TRANS2_CREATE_DIRECTORY,
			expected: "CREATE_DIRECTORY",
		},
		{
			name:     "SESSION_SETUP",
			subCmd:   smb_v10.TRANS2_SESSION_SETUP,
			expected: "SESSION_SETUP",
		},
		{
			name:     "GET_DFS_REFERRAL",
			subCmd:   smb_v10.TRANS2_GET_DFS_REFERRAL,
			expected: "GET_DFS_REFERRAL",
		},
		{
			name:     "REPORT_DFS_INCONSISTENCY",
			subCmd:   smb_v10.TRANS2_REPORT_DFS_INCONSISTENCY,
			expected: "REPORT_DFS_INCONSISTENCY",
		},
		{
			name:     "Unknown subcommand",
			subCmd:   smb_v10.Transaction2Subcommand(0xFFFF),
			expected: "UNKNOWN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.subCmd.String()
			if result != tt.expected {
				t.Errorf("String() = %v, want %v", result, tt.expected)
			}
		})
	}
}
