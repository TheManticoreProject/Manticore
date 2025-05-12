package guid

import (
	"testing"
)

func TestToBytes(t *testing.T) {
	guid := &GUID{A: 0x12345678, B: 0x1234, C: 0x5678, D: 0x9abc, E: 0xdef012345678}
	expected := []byte{
		0x78, 0x56, 0x34, 0x12,
		0x34, 0x12,
		0x78, 0x56,
		0x9a, 0xbc,
		0xde, 0xf0, 0x12, 0x34, 0x56, 0x78,
	}
	result := guid.ToBytes()
	for i, b := range result {
		if b != expected[i] {
			t.Errorf("Expected byte %x at position %d, but got %x", expected[i], i, b)
		}
	}
}

func TestInvolution(t *testing.T) {
	originalGuid := NewGUID()
	data := originalGuid.ToBytes()
	guid := &GUID{}
	guid.FromRawBytes(data)
	if !guid.Equal(originalGuid) {
		t.Errorf("GUIDs are not equal after involution. Before: %s, After: %s", originalGuid.ToFormatB(), guid.ToFormatB())
	}
}

func TestToFormatN(t *testing.T) {
	guid := &GUID{A: 0x12345678, B: 0x1234, C: 0x5678, D: 0x9abc, E: 0xdef012345678}
	expected := "12345678123456789abcdef012345678"
	result := guid.ToFormatN()
	if result != expected {
		t.Errorf("Expected %s, but got %s", expected, result)
	}
}

func TestToFormatD(t *testing.T) {
	guid := &GUID{A: 0x12345678, B: 0x1234, C: 0x5678, D: 0x9abc, E: 0xdef012345678}
	expected := "12345678-1234-5678-9abc-def012345678"
	result := guid.ToFormatD()
	if result != expected {
		t.Errorf("Expected %s, but got %s", expected, result)
	}
}

func TestToFormatB(t *testing.T) {
	guid := &GUID{A: 0x12345678, B: 0x1234, C: 0x5678, D: 0x9abc, E: 0xdef012345678}
	expected := "{12345678-1234-5678-9abc-def012345678}"
	result := guid.ToFormatB()
	if result != expected {
		t.Errorf("Expected %s, but got %s", expected, result)
	}
}

func TestToFormatP(t *testing.T) {
	guid := &GUID{A: 0x12345678, B: 0x1234, C: 0x5678, D: 0x9abc, E: 0xdef012345678}
	expected := "(12345678-1234-5678-9abc-def012345678)"
	result := guid.ToFormatP()
	if result != expected {
		t.Errorf("Expected %s, but got %s", expected, result)
	}
}

func TestToFormatX(t *testing.T) {
	guid := &GUID{A: 0x12345678, B: 0x1234, C: 0x5678, D: 0x9abc, E: 0xdef012345678}
	expected := "{0x12345678,0x1234,0x5678,{0x9a,0xbc,0xde,0xf0,0x12,0x34,0x56,0x78}}"
	result := guid.ToFormatX()
	if result != expected {
		t.Errorf("Expected %s, but got %s", expected, result)
	}
}

func TestFromStringToStringFormatN(t *testing.T) {
	data := "12345678123456789abcdef012345678"
	guid, err := FromString(data)
	if err != nil {
		t.Errorf("Error parsing GUID: %s", err)
	}
	if guid != nil {
		result := guid.ToFormatN()
		if result != guid.ToFormatN() {
			t.Errorf("Expected %s, but got %s", data, result)
		}
	}
}

func TestFromStringToStringFormatD(t *testing.T) {
	data := "12345678-1234-5678-9abc-def012345678"
	guid, err := FromString(data)
	if err != nil {
		t.Errorf("Error parsing GUID: %s", err)
	}
	if guid != nil {
		result := guid.ToFormatD()
		if result != guid.ToFormatD() {
			t.Errorf("Expected %s, but got %s", data, result)
		}
	}
}

func TestFromStringToStringFormatB(t *testing.T) {
	data := "{12345678-1234-5678-9abc-def012345678}"
	guid, err := FromString(data)
	if err != nil {
		t.Errorf("Error parsing GUID: %s", err)
	}
	if guid != nil {
		result := guid.ToFormatB()
		if result != guid.ToFormatB() {
			t.Errorf("Expected %s, but got %s", data, result)
		}
	}
}

func TestFromStringToStringFormatP(t *testing.T) {
	data := "(12345678-1234-5678-9abc-def012345678)"
	guid, err := FromString(data)
	if err != nil {
		t.Errorf("Error parsing GUID: %s", err)
	}
	if guid != nil {
		result := guid.ToFormatP()
		if result != guid.ToFormatP() {
			t.Errorf("Expected %s, but got %s", data, result)
		}
	}
}

func TestFromStringToStringFormatX(t *testing.T) {
	data := "{0x12345678,0x1234,0x5678,{0x9a,0xbc,0xde,0xf0,0x12,0x34,0x56,0x78}}"
	guid, err := FromString(data)
	if err != nil {
		t.Errorf("Error parsing GUID: %s", err)
	}
	if guid != nil {
		result := guid.ToFormatX()
		if result != guid.ToFormatX() {
			t.Errorf("Expected %s, but got %s", data, result)
		}
	}
}
