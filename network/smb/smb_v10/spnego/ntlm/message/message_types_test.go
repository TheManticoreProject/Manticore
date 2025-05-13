package message

import (
	"testing"
)

func TestMessageTypeString(t *testing.T) {
	tests := []struct {
		messageType MessageType
		expected    string
	}{
		{
			messageType: NTLM_NEGOTIATE,
			expected:    "NEGOTIATE",
		},
		{
			messageType: NTLM_CHALLENGE,
			expected:    "CHALLENGE",
		},
		{
			messageType: NTLM_AUTHENTICATE,
			expected:    "AUTHENTICATE",
		},
		{
			messageType: MessageType(0x00000004),
			expected:    "UNKNOWN_MESSAGE_TYPE(0x00000004)",
		},
	}

	for _, test := range tests {
		result := test.messageType.String()
		if result != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, result)
		}
	}
}
