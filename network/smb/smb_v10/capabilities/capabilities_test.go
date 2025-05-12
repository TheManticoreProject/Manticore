package capabilities_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/capabilities"
)

func TestCapabilitiesString(t *testing.T) {
	tests := []struct {
		name     string
		caps     capabilities.Capabilities
		expected string
	}{
		{
			name:     "No capabilities",
			caps:     capabilities.Capabilities(0),
			expected: "NONE",
		},
		{
			name:     "Single capability - CAP_RAW_MODE",
			caps:     capabilities.CAP_RAW_MODE,
			expected: "CAP_RAW_MODE",
		},
		{
			name:     "Single capability - CAP_UNICODE",
			caps:     capabilities.CAP_UNICODE,
			expected: "CAP_UNICODE",
		},
		{
			name:     "Multiple capabilities",
			caps:     capabilities.CAP_RAW_MODE | capabilities.CAP_UNICODE | capabilities.CAP_LARGE_FILES,
			expected: "CAP_LARGE_FILES|CAP_RAW_MODE|CAP_UNICODE",
		},
		{
			name: "All capabilities",
			caps: capabilities.CAP_RAW_MODE | capabilities.CAP_MPX_MODE | capabilities.CAP_UNICODE |
				capabilities.CAP_LARGE_FILES | capabilities.CAP_NT_SMBS | capabilities.CAP_RPC_REMOTE_APIS |
				capabilities.CAP_STATUS32 | capabilities.CAP_LEVEL_II_OPLOCKS | capabilities.CAP_LOCK_AND_READ |
				capabilities.CAP_NT_FIND | capabilities.CAP_DFS | capabilities.CAP_LARGE_READX,
			expected: "CAP_DFS|CAP_LARGE_FILES|CAP_LARGE_READX|CAP_LEVEL_II_OPLOCKS|CAP_LOCK_AND_READ|CAP_MPX_MODE|CAP_NT_FIND|CAP_NT_SMBS|CAP_RAW_MODE|CAP_RPC_REMOTE_APIS|CAP_STATUS32|CAP_UNICODE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.caps.String()
			if result != tt.expected {
				t.Errorf("Capabilities.String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCapabilitiesBitFlags(t *testing.T) {
	// Test that each capability constant has the correct bit value
	testCases := []struct {
		name     string
		cap      capabilities.Capabilities
		expected uint32
	}{
		{"CAP_RAW_MODE", capabilities.CAP_RAW_MODE, 0x00000001},
		{"CAP_MPX_MODE", capabilities.CAP_MPX_MODE, 0x00000002},
		{"CAP_UNICODE", capabilities.CAP_UNICODE, 0x00000004},
		{"CAP_LARGE_FILES", capabilities.CAP_LARGE_FILES, 0x00000008},
		{"CAP_NT_SMBS", capabilities.CAP_NT_SMBS, 0x00000010},
		{"CAP_RPC_REMOTE_APIS", capabilities.CAP_RPC_REMOTE_APIS, 0x00000020},
		{"CAP_STATUS32", capabilities.CAP_STATUS32, 0x00000040},
		{"CAP_LEVEL_II_OPLOCKS", capabilities.CAP_LEVEL_II_OPLOCKS, 0x00000080},
		{"CAP_LOCK_AND_READ", capabilities.CAP_LOCK_AND_READ, 0x00000100},
		{"CAP_NT_FIND", capabilities.CAP_NT_FIND, 0x00000200},
		{"CAP_DFS", capabilities.CAP_DFS, 0x00001000},
		{"CAP_LARGE_READX", capabilities.CAP_LARGE_READX, 0x00004000},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if uint32(tc.cap) != tc.expected {
				t.Errorf("%s = 0x%08X, want 0x%08X", tc.name, uint32(tc.cap), tc.expected)
			}
		})
	}
}
