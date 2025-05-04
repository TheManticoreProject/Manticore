package netbios_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/netbios"
)

func TestSessionMessageTypeString(t *testing.T) {
	tests := []struct {
		name     string
		message  netbios.SESSION_MESSAGE_TYPE
		expected string
	}{
		{
			name:     "SESSION_MESSAGE",
			message:  netbios.SESSION_MESSAGE,
			expected: "SESSION_MESSAGE",
		},
		{
			name:     "SESSION_REQUEST",
			message:  netbios.SESSION_REQUEST,
			expected: "SESSION_REQUEST",
		},
		{
			name:     "SESSION_POSITIVE_RESPONSE",
			message:  netbios.SESSION_POSITIVE_RESPONSE,
			expected: "SESSION_POSITIVE_RESPONSE",
		},
		{
			name:     "SESSION_NEGATIVE_RESPONSE",
			message:  netbios.SESSION_NEGATIVE_RESPONSE,
			expected: "SESSION_NEGATIVE_RESPONSE",
		},
		{
			name:     "SESSION_RETARGET_RESPONSE",
			message:  netbios.SESSION_RETARGET_RESPONSE,
			expected: "SESSION_RETARGET_RESPONSE",
		},
		{
			name:     "SESSION_KEEP_ALIVE",
			message:  netbios.SESSION_KEEP_ALIVE,
			expected: "SESSION_KEEP_ALIVE",
		},
		{
			name:     "UNKNOWN",
			message:  netbios.SESSION_MESSAGE_TYPE(0xFF),
			expected: "UNKNOWN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.message.String(); got != tt.expected {
				t.Errorf("SESSION_MESSAGE_TYPE.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}
