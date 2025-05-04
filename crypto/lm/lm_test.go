package lm

import (
	"testing"
)

func TestLMHash(t *testing.T) {
	tests := []struct {
		password string
		expected string
	}{
		{"cG9kYWxpcml1cwo", "350d1cb93bed51ccafe82d1cdb09b5de"},
		{"", "aad3b435b51404eeaad3b435b51404ee"},
		{"0123456", "5645f13f500882b2aad3b435b51404ee"},
		{"0123456789abcdef", "5645f13f500882b21ac3884b83324540"},
		{"0123456789abcde", "5645f13f500882b21ac3884b83324540"},
		{"0123456789abcd", "5645f13f500882b21ac3884b83324540"},
	}

	for _, test := range tests {
		t.Run(test.password, func(t *testing.T) {
			result := LMHash(test.password)
			if result != test.expected {
				t.Errorf("LMHash(%q) = %q; expected %q", test.password, result, test.expected)
			}
		})
	}
}
