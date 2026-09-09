package guid

import (
	"strings"
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

// TestNewGUIDIsRFC4122Version4 guards the defect: NewGUID previously left the
// version and variant bits as whatever the generator produced, so the value was
// not a valid version-4 GUID. RFC 4122 4.4 fixes six bits — the high nibble of
// time_hi_and_version and the two high bits of clock_seq_hi_and_reserved — and a
// peer that parses them (an SMB client reading a ServerGUID) can check both.
func TestNewGUIDIsRFC4122Version4(t *testing.T) {
	// Enough samples that a generator leaving these bits to chance fails
	// essentially always: one unfixed nibble alone passes only 1 time in 16.
	const samples = 256

	for i := 0; i < samples; i++ {
		g := NewGUID()
		if g == nil {
			t.Fatal("NewGUID() returned nil")
		}
		if version := (g.C >> 12) & 0xF; version != 4 {
			t.Fatalf("sample %d: version nibble = %d, want 4 (C = %#04x)", i, version, g.C)
		}
		if variant := (g.D >> 14) & 0x3; variant != 0x2 {
			t.Fatalf("sample %d: variant bits = %#b, want 0b10 (D = %#04x)", i, variant, g.D)
		}
		if g.E > 0xFFFFFFFFFFFF {
			t.Fatalf("sample %d: node field %#x exceeds 48 bits", i, g.E)
		}
	}
}

// TestNewGUIDIsDistinct checks that successive GUIDs differ. It does not prove the
// source is cryptographically secure — that cannot be asserted from output — but it
// does catch a generator that is unseeded or constant.
func TestNewGUIDIsDistinct(t *testing.T) {
	const samples = 1024

	seen := make(map[GUID]struct{}, samples)
	for i := 0; i < samples; i++ {
		g := *NewGUID()
		if _, dup := seen[g]; dup {
			t.Fatalf("NewGUID() returned a duplicate after %d samples: %s", i, g.ToFormatD())
		}
		seen[g] = struct{}{}
	}
}

// TestNewGUIDFormatsAsV4 checks that the generated value survives the string form,
// where the version digit is directly visible as the first character of the third
// group.
func TestNewGUIDFormatsAsV4(t *testing.T) {
	g := NewGUID()
	formatted := g.ToFormatD()

	groups := strings.Split(formatted, "-")
	if len(groups) != 5 {
		t.Fatalf("ToFormatD() = %q, want five hyphen-separated groups", formatted)
	}
	if groups[2][0] != '4' {
		t.Errorf("third group of %q starts with %q, want '4'", formatted, groups[2][0])
	}
	if c := groups[3][0]; c != '8' && c != '9' && c != 'a' && c != 'b' {
		t.Errorf("fourth group of %q starts with %q, want one of 8/9/a/b", formatted, c)
	}
}
