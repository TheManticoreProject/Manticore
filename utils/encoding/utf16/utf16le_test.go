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
