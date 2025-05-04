package key_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/keycredential/key"
)

func TestKeyUsage_Parse(t *testing.T) {
	tests := []struct {
		input    byte
		expected key.KeyUsage
	}{
		{0x00, key.KeyUsage{Value: 0x00, RawBytes: []byte{0x00}, RawBytesSize: 1}},
		{0x01, key.KeyUsage{Value: 0x01, RawBytes: []byte{0x01}, RawBytesSize: 1}},
		{0x02, key.KeyUsage{Value: 0x02, RawBytes: []byte{0x02}, RawBytesSize: 1}},
		{0x03, key.KeyUsage{Value: 0x03, RawBytes: []byte{0x03}, RawBytesSize: 1}},
		{0x04, key.KeyUsage{Value: 0x04, RawBytes: []byte{0x04}, RawBytesSize: 1}},
		{0x07, key.KeyUsage{Value: 0x07, RawBytes: []byte{0x07}, RawBytesSize: 1}},
		{0x08, key.KeyUsage{Value: 0x08, RawBytes: []byte{0x08}, RawBytesSize: 1}},
		{0x09, key.KeyUsage{Value: 0x09, RawBytes: []byte{0x09}, RawBytesSize: 1}},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("input=%x", tt.input), func(t *testing.T) {
			var ku key.KeyUsage
			ku.FromBytes(tt.input)
			if ku.Value != tt.expected.Value || !bytes.Equal(ku.RawBytes, tt.expected.RawBytes) || ku.RawBytesSize != tt.expected.RawBytesSize {
				t.Errorf("got %+v, want %+v", ku, tt.expected)
			}
		})
	}
}
