package header_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/llmnr/message/header"
)

func TestIsQueryAndIsResponse(t *testing.T) {
	var flag header.Flags

	flag = 0 // QR not set
	if !flag.IsQuery() {
		t.Errorf("Flags(0) should be query (IsQuery == true)")
	}
	if flag.IsResponse() {
		t.Errorf("Flags(0) should not be response (IsResponse == false)")
	}

	flag = header.FlagQR // QR set
	if flag.IsQuery() {
		t.Errorf("Flags(QR) should not be query (IsQuery == false)")
	}
	if !flag.IsResponse() {
		t.Errorf("Flags(QR) should be response (IsResponse == true)")
	}
}

func TestIsConflict(t *testing.T) {
	var f header.Flags = 0
	if f.IsConflict() {
		t.Error("Flags(0) should not have conflict")
	}
	f = header.FlagC
	if !f.IsConflict() {
		t.Error("Flags(C) should have conflict")
	}
}

func TestIsTruncation(t *testing.T) {
	var f header.Flags = 0
	if f.IsTruncation() {
		t.Error("Flags(0) should not have truncation")
	}
	f = header.FlagTC
	if !f.IsTruncation() {
		t.Error("Flags(TC) should have truncation")
	}
}

func TestIsTentative(t *testing.T) {
	var f header.Flags = 0
	if f.IsTentative() {
		t.Error("Flags(0) should not be tentative")
	}
	f = header.FlagT
	if !f.IsTentative() {
		t.Error("Flags(T) should be tentative")
	}
}

func TestFlagsMarshalUnmarshal(t *testing.T) {
	testVals := []header.Flags{
		0,
		header.FlagQR,
		header.FlagC,
		header.FlagTC,
		header.FlagT,
		header.FlagQR | header.FlagC,
		header.FlagQR | header.FlagTC | header.FlagT,
		header.FlagQR | header.MaskOpcode | header.FlagC | header.FlagTC | header.FlagT,
	}
	for _, orig := range testVals {
		cpy := orig
		b, err := cpy.Marshal()
		if err != nil {
			t.Fatalf("Marshal(%#v) failed: %v", orig, err)
		}
		if len(b) != 2 {
			t.Errorf("Marshal(%#v) = %v (len %d), want 2 bytes", orig, b, len(b))
		}

		var decoded header.Flags
		n, err := decoded.Unmarshal(b)
		if err != nil {
			t.Fatalf("Unmarshal(%v) failed: %v", b, err)
		}
		if n != 2 {
			t.Errorf("Unmarshal() read %d bytes; want 2", n)
		}
		if decoded != orig {
			t.Errorf("Unmarshal(Marshal(%#v)) = %#v; want %#v", orig, decoded, orig)
		}
	}
}

func TestFlagsUnmarshalInvalidLength(t *testing.T) {
	var f header.Flags
	_, err := f.Unmarshal([]byte{1})
	if err == nil {
		t.Error("Unmarshal should fail on input of 1 byte (should error), got nil error")
	}
	_, err = f.Unmarshal([]byte{0, 1, 2})
	if err == nil {
		t.Error("Unmarshal should fail on input of 3 bytes (should error), got nil error")
	}
}

func TestFlagsString(t *testing.T) {
	tests := []struct {
		flags    header.Flags
		expected string
	}{
		// Query messages (QR bit clear) must not emit "QR".
		{0, ""},
		{header.FlagC, "C"},
		{header.FlagTC, "TC"},
		{header.FlagT, "T"},
		// Response messages (QR bit set) emit "QR".
		{header.FlagQR, "QR"},
		{header.FlagQR | header.FlagC, "QR|C"},
		{header.FlagQR | header.FlagTC | header.FlagT, "QR|TC|T"},
		// The Opcode and RCODE fields are 4-bit values, rendered only when non-zero
		// and positioned relative to the single-bit flags in the same order as the wire.
		{header.MaskOpcode | header.FlagC | header.FlagTC | header.FlagT, "OPCODE=15|C|TC|T"},
		{header.FlagQR | header.MaskRCODE, "QR|RCODE=15"},
	}
	for _, tt := range tests {
		got := tt.flags.String()
		if got != tt.expected {
			t.Errorf("Flags.String() = %q, want %q for value %#x", got, tt.expected, uint16(tt.flags))
		}
	}
}

// TestFlagBitPositions is a known-answer test pinning the exact 16-bit word
// produced when each flag or field is set individually, per the RFC 4795
// §2.1.1 header layout (QR bit0, Opcode bits1-4, C bit5, TC bit6, T bit7,
// Z bits8-11, RCODE bits12-15). These assertions guard against regressions in
// the bit positions themselves.
func TestFlagBitPositions(t *testing.T) {
	tests := []struct {
		name string
		flag header.Flags
		want uint16
	}{
		{"QR", header.FlagQR, 0x8000},
		{"C", header.FlagC, 0x0400},
		{"TC", header.FlagTC, 0x0200},
		{"T", header.FlagT, 0x0100},
		{"MaskOpcode", header.MaskOpcode, 0x7800},
		{"MaskZ", header.MaskZ, 0x00F0},
		{"MaskRCODE", header.MaskRCODE, 0x000F},
	}
	for _, tt := range tests {
		if got := uint16(tt.flag); got != tt.want {
			t.Errorf("%s = %#04x, want %#04x", tt.name, got, tt.want)
		}
	}
}

// TestOpcodeAccessors verifies the 4-bit Opcode field (bits 1-4) is read and
// written at the correct position and round-trips. Opcode is a normal integer
// occupying bits 1-4 (shift 11), so Opcode=1 => 0x0800 and the maximum value
// 15 fills the whole field (0x7800).
func TestOpcodeAccessors(t *testing.T) {
	tests := []struct {
		opcode uint8
		want   uint16
	}{
		{0, 0x0000},
		{1, 0x0800},
		{2, 0x1000},
		{15, 0x7800},
	}
	for _, tt := range tests {
		var f header.Flags
		f.SetOpcode(tt.opcode)
		if uint16(f) != tt.want {
			t.Errorf("SetOpcode(%d) = %#04x, want %#04x", tt.opcode, uint16(f), tt.want)
		}
		if f.Opcode() != tt.opcode {
			t.Errorf("Opcode() after SetOpcode(%d) = %d, want %d", tt.opcode, f.Opcode(), tt.opcode)
		}
	}

	// Setting the Opcode must not disturb neighbouring bits (QR / C / RCODE).
	f := header.FlagQR | header.FlagC | header.MaskRCODE
	f.SetOpcode(5)
	if !f.IsResponse() || !f.IsConflict() || f.RCODE() != 15 || f.Opcode() != 5 {
		t.Errorf("SetOpcode clobbered neighbouring fields: %#04x", uint16(f))
	}
}

// TestRCODEAccessors verifies the 4-bit RCODE field (bits 12-15) is read and
// written at the correct position and round-trips. RCODE is the low nibble, so
// RCODE=3 => 0x0003.
func TestRCODEAccessors(t *testing.T) {
	tests := []struct {
		rcode uint8
		want  uint16
	}{
		{0, 0x0000},
		{3, 0x0003},
		{15, 0x000F},
	}
	for _, tt := range tests {
		var f header.Flags
		f.SetRCODE(tt.rcode)
		if uint16(f) != tt.want {
			t.Errorf("SetRCODE(%d) = %#04x, want %#04x", tt.rcode, uint16(f), tt.want)
		}
		if f.RCODE() != tt.rcode {
			t.Errorf("RCODE() after SetRCODE(%d) = %d, want %d", tt.rcode, f.RCODE(), tt.rcode)
		}
	}
}

// TestZAccessors verifies the 4-bit Z reserved field (bits 8-11) is read and
// written at the correct position and round-trips.
func TestZAccessors(t *testing.T) {
	tests := []struct {
		z    uint8
		want uint16
	}{
		{0, 0x0000},
		{1, 0x0010},
		{15, 0x00F0},
	}
	for _, tt := range tests {
		var f header.Flags
		f.SetZ(tt.z)
		if uint16(f) != tt.want {
			t.Errorf("SetZ(%d) = %#04x, want %#04x", tt.z, uint16(f), tt.want)
		}
		if f.Z() != tt.z {
			t.Errorf("Z() after SetZ(%d) = %d, want %d", tt.z, f.Z(), tt.z)
		}
	}
}
