package subcommands_test

import (
	"testing"

	smb_v10 "github.com/TheManticoreProject/Manticore/network/smb/smb_v10/subcommands"
)

func TestNtTransactSubcommandString(t *testing.T) {
	tests := []struct {
		name     string
		subCmd   smb_v10.NtTransactSubcommand
		expected string
	}{
		{
			name:     "CREATE",
			subCmd:   smb_v10.NT_TRANSACT_CREATE,
			expected: "CREATE",
		},
		{
			name:     "IOCTL",
			subCmd:   smb_v10.NT_TRANSACT_IOCTL,
			expected: "IOCTL",
		},
		{
			name:     "SET_SECURITY_DESC",
			subCmd:   smb_v10.NT_TRANSACT_SET_SECURITY_DESC,
			expected: "SET_SECURITY_DESC",
		},
		{
			name:     "NOTIFY_CHANGE",
			subCmd:   smb_v10.NT_TRANSACT_NOTIFY_CHANGE,
			expected: "NOTIFY_CHANGE",
		},
		{
			name:     "RENAME",
			subCmd:   smb_v10.NT_TRANSACT_RENAME,
			expected: "RENAME",
		},
		{
			name:     "QUERY_SECURITY_DESC",
			subCmd:   smb_v10.NT_TRANSACT_QUERY_SECURITY_DESC,
			expected: "QUERY_SECURITY_DESC",
		},
		{
			name:     "QUERY_QUOTA",
			subCmd:   smb_v10.NT_TRANSACT_QUERY_QUOTA,
			expected: "QUERY_QUOTA",
		},
		{
			name:     "SET_QUOTA",
			subCmd:   smb_v10.NT_TRANSACT_SET_QUOTA,
			expected: "SET_QUOTA",
		},
		{
			name:     "Unknown subcommand",
			subCmd:   smb_v10.NtTransactSubcommand(0xFFFF),
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
