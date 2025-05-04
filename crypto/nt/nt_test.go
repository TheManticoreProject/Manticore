package nt

import (
	"testing"
)

func TestNTHash(t *testing.T) {
	tests := []struct {
		password string
		expected string
	}{
		{"cG9kYWxpcml1cwo", "625d671c4b67f53df60c694e690bf5ac"},
		{"00000000000000000000000000000000", "312b863247fa4d3fbba0969e4acb7682"},
		{"", "31d6cfe0d16ae931b73c59d7e0c089c0"},
	}

	for _, test := range tests {
		result := NTHashHex(test.password)
		if result != test.expected {
			t.Errorf("NTHash(%q) = %q; expected %q", test.password, result, test.expected)
		}
	}
}
