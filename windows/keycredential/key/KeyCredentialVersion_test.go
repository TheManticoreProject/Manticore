package key_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/keycredential/key"
)

func TestKeyCredentialVersionFromBytes(t *testing.T) {
	tests := []struct {
		input    []byte
		expected uint32
	}{
		{[]byte{0x00, 0x00, 0x00, 0x00}, key.KeyCredentialVersion_0},
		{[]byte{0x00, 0x01, 0x00, 0x00}, key.KeyCredentialVersion_1},
		{[]byte{0x00, 0x02, 0x00, 0x00}, key.KeyCredentialVersion_2},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("input: %v", test.input), func(t *testing.T) {
			var kcv key.KeyCredentialVersion
			kcv.FromBytes(test.input)
			if kcv.Value != test.expected {
				t.Errorf("Expected %d, but got %d", test.expected, kcv.Value)
			}
		})
	}
}

func TestKeyCredentialVersionString(t *testing.T) {
	tests := []struct {
		version  key.KeyCredentialVersion
		expected string
	}{
		{key.KeyCredentialVersion{Value: key.KeyCredentialVersion_0}, "KeyCredential_v0"},
		{key.KeyCredentialVersion{Value: key.KeyCredentialVersion_1}, "KeyCredential_v1"},
		{key.KeyCredentialVersion{Value: key.KeyCredentialVersion_2}, "KeyCredential_v2"},
		{key.KeyCredentialVersion{Value: 0x00000300}, "Unknown version: 768"},
	}

	for _, test := range tests {
		t.Run(test.expected, func(t *testing.T) {
			result := test.version.String()
			if result != test.expected {
				t.Errorf("Expected %s, but got %s", test.expected, result)
			}
		})
	}
}

func TestKeyCredentialVersionToBytes(t *testing.T) {
	tests := []struct {
		version  key.KeyCredentialVersion
		expected []byte
	}{
		{key.KeyCredentialVersion{Value: key.KeyCredentialVersion_0}, []byte{0x00, 0x00, 0x00, 0x00}},
		{key.KeyCredentialVersion{Value: key.KeyCredentialVersion_1}, []byte{0x00, 0x01, 0x00, 0x00}},
		{key.KeyCredentialVersion{Value: key.KeyCredentialVersion_2}, []byte{0x00, 0x02, 0x00, 0x00}},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("version: %d", test.version.Value), func(t *testing.T) {
			result := test.version.ToBytes()
			if !bytes.Equal(result, test.expected) {
				t.Errorf("Expected %v, but got %v", test.expected, result)
			}
		})
	}
}
