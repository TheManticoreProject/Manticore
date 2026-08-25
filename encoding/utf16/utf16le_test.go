package utf16

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestEncodeUTF16LE(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "680065006c006c006f00"},
	}

	for _, test := range tests {
		result := EncodeUTF16LE(test.input)
		rawExpected, err := hex.DecodeString(test.expected)
		if err != nil {
			t.Errorf("Could not decode hex string: %q", test.expected)
		}
		if !bytes.Equal(result, rawExpected) {
			t.Errorf("EncodeUTF16LE(%q) = %q; expected %q", test.input, result, rawExpected)
		}
	}
}

func TestDecodeUTF16LE(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"680065006c006c006f00", "hello"},
	}

	for _, test := range tests {
		rawUTF16LE, err := hex.DecodeString(test.input)
		if err != nil {
			t.Errorf("Could not decode hex string: %q", test.input)
		}
		result := DecodeUTF16LE(rawUTF16LE)
		if result != test.expected {
			t.Errorf("DecodeUTF16LE(%q) = %q; expected %q", test.input, result, rawUTF16LE)
		}
	}
}

func TestIsUTF16LE(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"0065006c006c006f0068", false},
		{"680065006c006c006f00", true},
	}

	for _, test := range tests {
		rawUTF16LE, err := hex.DecodeString(test.input)
		if err != nil {
			t.Errorf("Could not decode hex string: %q", test.input)
		}
		result := IsUTF16LE(rawUTF16LE)
		if result != test.expected {
			t.Errorf("IsUTF16LE(%q) = %v; expected %v", test.input, result, test.expected)
		}
	}
}

// TestDecodeUTF16LEOddLength asserts an odd-length input drops its trailing byte
// rather than panicking.
//
// The decoder is reached from fields taken off the network, where the length is
// whatever the sender said: a client that negotiated Unicode and sent a name whose
// terminator was found at an odd offset produced a one-byte slice here, and the
// panic that followed killed the connection serving it.
func TestDecodeUTF16LEOddLength(t *testing.T) {
	cases := []struct {
		name  string
		input []byte
		want  string
	}{
		{"a single byte", []byte{0x5C}, ""},
		{"one character and a half", []byte{0x41, 0x00, 0x42}, "A"},
		{"three and a half", []byte{0x41, 0x00, 0x42, 0x00, 0x43, 0x00, 0x44}, "ABC"},
		{"empty", []byte{}, ""},
		{"nil", nil, ""},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("DecodeUTF16LE(% x) panicked: %v", tc.input, r)
				}
			}()
			if got := DecodeUTF16LE(tc.input); got != tc.want {
				t.Fatalf("DecodeUTF16LE(% x) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}
