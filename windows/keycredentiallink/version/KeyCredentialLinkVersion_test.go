package version_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/keycredentiallink/version"
)

func TestKeyCredentialLinkVersion_Unmarshal(t *testing.T) {
	tests := []struct {
		input    []byte
		expected version.KeyCredentialLinkVersion
	}{
		{[]byte{0x00, 0x00, 0x00, 0x00}, version.KeyCredentialLinkVersion{Value: version.KeyCredentialLinkVersion_0}},
		{[]byte{0x00, 0x01, 0x00, 0x00}, version.KeyCredentialLinkVersion{Value: version.KeyCredentialLinkVersion_1}},
		{[]byte{0x00, 0x02, 0x00, 0x00}, version.KeyCredentialLinkVersion{Value: version.KeyCredentialLinkVersion_2}},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("input: %v", test.input), func(t *testing.T) {
			var kcv version.KeyCredentialLinkVersion
			_, err := kcv.Unmarshal(test.input)
			if err != nil {
				t.Errorf("\n| Expected : no error \n| But got  : %v", err)
			}
			if kcv.Value != test.expected.Value {
				t.Errorf("\n| Expected : 0x%08X \n| But got  : 0x%08X", test.expected.Value, kcv.Value)
			}
		})
	}
}

func TestKeyCredentialLinkVersion_String(t *testing.T) {
	tests := []struct {
		version  version.KeyCredentialLinkVersion
		expected string
	}{
		{version.KeyCredentialLinkVersion{Value: version.KeyCredentialLinkVersion_0}, "KeyCredentialLink_v0"},
		{version.KeyCredentialLinkVersion{Value: version.KeyCredentialLinkVersion_1}, "KeyCredentialLink_v1"},
		{version.KeyCredentialLinkVersion{Value: version.KeyCredentialLinkVersion_2}, "KeyCredentialLink_v2"},
		{version.KeyCredentialLinkVersion{Value: 0x00000300}, "Unknown version: 0x00000300"},
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

func TestKeyCredentialLinkVersion_Marshal(t *testing.T) {
	tests := []struct {
		version  version.KeyCredentialLinkVersion
		expected []byte
	}{
		{version.KeyCredentialLinkVersion{Value: version.KeyCredentialLinkVersion_0}, []byte{0x00, 0x00, 0x00, 0x00}},
		{version.KeyCredentialLinkVersion{Value: version.KeyCredentialLinkVersion_1}, []byte{0x00, 0x01, 0x00, 0x00}},
		{version.KeyCredentialLinkVersion{Value: version.KeyCredentialLinkVersion_2}, []byte{0x00, 0x02, 0x00, 0x00}},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("version: %d", test.version.Value), func(t *testing.T) {
			result, err := test.version.Marshal()
			if err != nil {
				t.Errorf("Expected no error, but got %v", err)
			}
			if !bytes.Equal(result, test.expected) {
				t.Errorf("Expected %v, but got %v", test.expected, result)
			}
		})
	}
}

func TestKeyCredentialLinkVersion_Unmarshal_TooShort(t *testing.T) {
	inputs := [][]byte{
		nil,
		{},
		{0x00},
		{0x00, 0x02},
		{0x00, 0x02, 0x00},
	}

	for _, input := range inputs {
		t.Run(fmt.Sprintf("len=%d", len(input)), func(t *testing.T) {
			var kcv version.KeyCredentialLinkVersion
			bytesRead, err := kcv.Unmarshal(input)
			if err == nil {
				t.Errorf("Expected an error for a %d-byte input, but got nil", len(input))
			}
			if bytesRead != 0 {
				t.Errorf("Expected 0 bytes read, but got %d", bytesRead)
			}
		})
	}
}
