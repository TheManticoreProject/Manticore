package strength_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/keycredentiallink/key/strength"
)

func TestKeyStrength_Unmarshal_Valid(t *testing.T) {
	tests := []struct {
		input    []byte
		expected strength.KeyStrength
		name     string
	}{
		// KeyStrength is one byte ([MS-ADTS] 2.2.20.2.1); trailing bytes belong
		// to the next field and must not be consumed.
		{[]byte{0x00, 0xAA, 0xBB, 0xCC}, strength.KeyStrength_Unknown, "Unknown"},
		{[]byte{0x01, 0xAA, 0xBB, 0xCC}, strength.KeyStrength_Weak, "Weak"},
		{[]byte{0x02}, strength.KeyStrength_Normal, "Normal"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var ks strength.KeyStrength
			bytesRead, err := ks.Unmarshal(test.input)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if bytesRead != 1 {
				t.Errorf("Expected to read 1 byte, but read %d", bytesRead)
			}
			if uint32(ks) != uint32(test.expected) {
				t.Errorf("Expected value %d, but got %d", uint32(test.expected), uint32(ks))
			}
			if ks.String() != test.name {
				t.Errorf("Expected name %s, but got %s", test.name, ks.String())
			}
		})
	}
}

func TestKeyStrength_Unmarshal_UnknownValue(t *testing.T) {
	// Value 0xFF is not defined; expect String() to fall back to "Unknown"
	input := []byte{0xFF}
	var ks strength.KeyStrength
	bytesRead, err := ks.Unmarshal(input)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if bytesRead != 1 {
		t.Fatalf("Expected to read 1 byte, but read %d", bytesRead)
	}
	if uint32(ks) != 0xFF {
		t.Fatalf("Expected value 255, but got %d", uint32(ks))
	}
	if ks.String() != "Unknown" {
		t.Fatalf("Expected name %q for unknown value, but got %q", "Unknown", ks.String())
	}
}

func TestKeyStrength_Unmarshal_ErrorTooShort(t *testing.T) {
	var ks strength.KeyStrength
	if _, err := ks.Unmarshal(nil); err == nil {
		t.Fatalf("expected error for too short input, got nil")
	}
}

func TestKeyStrength_Marshal_Valid(t *testing.T) {
	tests := []struct {
		value    strength.KeyStrength
		expected []byte
	}{
		{strength.KeyStrength_Unknown, []byte{0x00}},
		{strength.KeyStrength_Weak, []byte{0x01}},
		{strength.KeyStrength_Normal, []byte{0x02}},
	}

	for _, test := range tests {
		t.Run(test.value.String(), func(t *testing.T) {
			ks := test.value
			result, err := ks.Marshal()
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if len(result) != len(test.expected) {
				t.Errorf("Expected %d bytes, but got %d", len(test.expected), len(result))
			}
			for i := range result {
				if result[i] != test.expected[i] {
					t.Errorf("Byte %d: expected %02x, but got %02x", i, test.expected[i], result[i])
				}
			}
		})
	}
}

func TestKeyStrength_Marshal_RejectsOverflow(t *testing.T) {
	ks := strength.KeyStrength(0x100)
	if _, err := ks.Marshal(); err == nil {
		t.Fatal("a value that does not fit the 1-byte field was marshalled")
	}
}
