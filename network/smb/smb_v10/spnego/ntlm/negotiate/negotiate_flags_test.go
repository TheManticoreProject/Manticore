package negotiate_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/spnego/ntlm/negotiate"
)

func TestNegotiateFlagsString(t *testing.T) {
	tests := []struct {
		flags    negotiate.NegotiateFlags
		expected string
	}{
		{
			flags:    negotiate.NTLMSSP_NEGOTIATE_UNICODE,
			expected: "NTLMSSP_NEGOTIATE_UNICODE",
		},
		{
			flags:    negotiate.NTLMSSP_NEGOTIATE_UNICODE | negotiate.NTLMSSP_NEGOTIATE_OEM,
			expected: "NTLMSSP_NEGOTIATE_UNICODE|NTLMSSP_NEGOTIATE_OEM",
		},
		{
			flags:    negotiate.NTLMSSP_NEGOTIATE_UNICODE | negotiate.NTLMSSP_NEGOTIATE_OEM | negotiate.NTLMSSP_REQUEST_TARGET,
			expected: "NTLMSSP_NEGOTIATE_UNICODE|NTLMSSP_NEGOTIATE_OEM|NTLMSSP_REQUEST_TARGET",
		},
	}

	for _, test := range tests {
		result := test.flags.String()
		if result != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, result)
		}
	}
}

func TestNegotiateFlagsHas(t *testing.T) {
	tests := []struct {
		flags    negotiate.NegotiateFlags
		check    negotiate.NegotiateFlags
		expected bool
	}{
		{
			flags:    negotiate.NTLMSSP_NEGOTIATE_UNICODE,
			check:    negotiate.NTLMSSP_NEGOTIATE_UNICODE,
			expected: true,
		},
		{
			flags:    negotiate.NTLMSSP_NEGOTIATE_UNICODE | negotiate.NTLMSSP_NEGOTIATE_OEM,
			check:    negotiate.NTLMSSP_NEGOTIATE_OEM,
			expected: true,
		},
		{
			flags:    negotiate.NTLMSSP_NEGOTIATE_UNICODE,
			check:    negotiate.NTLMSSP_NEGOTIATE_OEM,
			expected: false,
		},
	}

	for _, test := range tests {
		result := test.flags.Has(test.check)
		if result != test.expected {
			t.Errorf("Expected %v, got %v", test.expected, result)
		}
	}
}
