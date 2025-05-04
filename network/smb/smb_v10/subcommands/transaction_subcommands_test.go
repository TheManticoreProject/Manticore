package subcommands_test

import (
	"testing"

	smb_v10 "github.com/TheManticoreProject/Manticore/network/smb/smb_v10/subcommands"
)

func TestTransactionSubcommandString(t *testing.T) {
	tests := []struct {
		name     string
		subCmd   smb_v10.TransactionSubcommand
		expected string
	}{
		{
			name:     "SET_NMPIPE_STATE",
			subCmd:   smb_v10.TRANS_SET_NMPIPE_STATE,
			expected: "SET_NMPIPE_STATE",
		},
		{
			name:     "RAW_READ_NMPIPE",
			subCmd:   smb_v10.TRANS_RAW_READ_NMPIPE,
			expected: "RAW_READ_NMPIPE",
		},
		{
			name:     "QUERY_NMPIPE_STATE",
			subCmd:   smb_v10.TRANS_QUERY_NMPIPE_STATE,
			expected: "QUERY_NMPIPE_STATE",
		},
		{
			name:     "QUERY_NMPIPE_INFO",
			subCmd:   smb_v10.TRANS_QUERY_NMPIPE_INFO,
			expected: "QUERY_NMPIPE_INFO",
		},
		{
			name:     "PEEK_NMPIPE",
			subCmd:   smb_v10.TRANS_PEEK_NMPIPE,
			expected: "PEEK_NMPIPE",
		},
		{
			name:     "TRANSACT_NMPIPE",
			subCmd:   smb_v10.TRANS_TRANSACT_NMPIPE,
			expected: "TRANSACT_NMPIPE",
		},
		{
			name:     "RAW_WRITE_NMPIPE",
			subCmd:   smb_v10.TRANS_RAW_WRITE_NMPIPE,
			expected: "RAW_WRITE_NMPIPE",
		},
		{
			name:     "READ_NMPIPE",
			subCmd:   smb_v10.TRANS_READ_NMPIPE,
			expected: "READ_NMPIPE",
		},
		{
			name:     "WRITE_NMPIPE",
			subCmd:   smb_v10.TRANS_WRITE_NMPIPE,
			expected: "WRITE_NMPIPE",
		},
		{
			name:     "WAIT_NMPIPE",
			subCmd:   smb_v10.TRANS_WAIT_NMPIPE,
			expected: "WAIT_NMPIPE",
		},
		{
			name:     "CALL_NMPIPE",
			subCmd:   smb_v10.TRANS_CALL_NMPIPE,
			expected: "CALL_NMPIPE",
		},
		{
			name:     "Unknown subcommand",
			subCmd:   smb_v10.TransactionSubcommand(0xFFFF),
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
