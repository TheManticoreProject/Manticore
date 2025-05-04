package dcc

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/crypto/nt"
)

func TestDCCHash(t *testing.T) {
	tests := []struct {
		password string
		username string
		expected string
	}{
		{"cG9kYWxpcml1cwo", "podalirius", "72c257413870227435b474b3542a5c4b:podalirius"},
	}

	for _, test := range tests {
		result := DCCHashFromPasswordToHashcatString(test.password, test.username)
		if result != test.expected {
			t.Errorf("DCCHashFromPasswordToHashcatString(%q, %q) = %q; expected %q", test.password, test.username, result, test.expected)
		}
	}
}

func TestDCCHashFromPassword(t *testing.T) {
	tests := []struct {
		password string
		username string
		expected string
	}{
		{"cG9kYWxpcml1cwo", "podalirius", "72c257413870227435b474b3542a5c4b"},
	}

	for _, test := range tests {
		result := DCCHashFromPasswordToHex(test.password, test.username)
		if result != test.expected {
			t.Errorf("DCCHashFromPasswordToHex(%q, %q) = %q; expected %q", test.password, test.username, result, test.expected)
		}
	}
}

func TestDCCHashFromNTHash(t *testing.T) {
	tests := []struct {
		ntHash   [16]byte
		username string
		expected string
	}{
		{nt.NTHash("cG9kYWxpcml1cwo"), "podalirius", "72c257413870227435b474b3542a5c4b"},
	}

	for _, test := range tests {
		result := DCCHashFromNTHashToHex(test.ntHash, test.username)
		if result != test.expected {
			t.Errorf("DCCHashFromNTHashToHex(%q, %q) = %q; expected %q", test.ntHash, test.username, result, test.expected)
		}
	}
}

func TestDCCHashFromPasswordToHex(t *testing.T) {
	tests := []struct {
		password string
		username string
		expected string
	}{
		{"cG9kYWxpcml1cwo", "podalirius", "72c257413870227435b474b3542a5c4b"},
	}

	for _, test := range tests {
		result := DCCHashFromPasswordToHex(test.password, test.username)
		if result != test.expected {
			t.Errorf("DCCHashFromPasswordToHex(%q, %q) = %q; expected %q", test.password, test.username, result, test.expected)
		}
	}
}

func TestDCCHashFromNTHashToHex(t *testing.T) {
	tests := []struct {
		ntHash   [16]byte
		username string
		expected string
	}{
		{nt.NTHash("cG9kYWxpcml1cwo"), "podalirius", "72c257413870227435b474b3542a5c4b"},
	}

	for _, test := range tests {
		result := DCCHashFromNTHashToHex(test.ntHash, test.username)
		if result != test.expected {
			t.Errorf("DCCHashFromNTHashToHex(%q, %q) = %q; expected %q", test.ntHash, test.username, result, test.expected)
		}
	}
}

func TestDCCHashFromPasswordToHashcatString(t *testing.T) {
	tests := []struct {
		password string
		username string
		expected string
	}{
		{"cG9kYWxpcml1cwo", "podalirius", "72c257413870227435b474b3542a5c4b:podalirius"},
	}

	for _, test := range tests {
		result := DCCHashFromPasswordToHashcatString(test.password, test.username)
		if result != test.expected {
			t.Errorf("DCCHashFromPasswordToHashcatString(%q, %q) = %q; expected %q", test.password, test.username, result, test.expected)
		}
	}
}

func TestDCCHashFromNTHashToHashcatString(t *testing.T) {
	tests := []struct {
		ntHash   [16]byte
		username string
		expected string
	}{
		{nt.NTHash("cG9kYWxpcml1cwo"), "podalirius", "72c257413870227435b474b3542a5c4b:podalirius"},
	}

	for _, test := range tests {
		result := DCCHashFromNTHashToHashcatString(test.ntHash, test.username)
		if result != test.expected {
			t.Errorf("DCCHashFromNTHashToHashcatString(%q, %q) = %q; expected %q", test.ntHash, test.username, result, test.expected)
		}
	}
}
