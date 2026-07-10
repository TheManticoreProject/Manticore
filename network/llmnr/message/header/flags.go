package header

import (
	"encoding/binary"
	"fmt"
	"strings"
)

type Flags uint16

// LLMNR Header Flags and fields.
//
// Per RFC 4795 §2.1.1, the second 16-bit word of the LLMNR header reuses the
// DNS header layout (RFC 1035 §4.1.1) with LLMNR-specific bits. Using DNS bit
// numbering where bit 0 is the most significant bit of the 16-bit word:
//
//		 0   1   2   3   4   5   6   7   8   9  10  11  12  13  14  15
//		+--+----+----+----+----+--+--+--+--+--+--+--+----+----+----+----+
//		|QR|         Opcode        | C|TC| T|      Z    |      RCODE     |
//		+--+----+----+----+----+--+--+--+--+--+--+--+----+----+----+----+
//
//	  - QR      : bit 0             -> 1 << 15
//	  - Opcode  : bits 1-4  (4-bit) -> mask 0x7800, shift 11
//	  - C       : bit 5             -> 1 << 10
//	  - TC      : bit 6             -> 1 << 9
//	  - T       : bit 7             -> 1 << 8
//	  - Z       : bits 8-11 (4-bit) -> mask 0x00F0, shift 4 (reserved, MUST be 0)
//	  - RCODE   : bits 12-15 (4-bit) -> mask 0x000F
const (
	FlagQR Flags = 1 << 15 // Query/Response flag (bit 0)
	FlagC  Flags = 1 << 10 // Conflict flag (bit 5)
	FlagTC Flags = 1 << 9  // Truncation flag (bit 6)
	FlagT  Flags = 1 << 8  // Tentative flag (bit 7)

	// Field masks and shifts for the multi-bit Opcode, Z and RCODE fields.
	MaskOpcode  Flags = 0x7800 // Opcode field (bits 1-4)
	ShiftOpcode uint  = 11
	MaskZ       Flags = 0x00F0 // Z reserved field (bits 8-11)
	ShiftZ      uint  = 4
	MaskRCODE   Flags = 0x000F // RCODE field (bits 12-15)
)

// IsQuery returns true if the flags are set for a query.
func (f Flags) IsQuery() bool {
	return f&FlagQR == 0
}

// IsResponse returns true if the flags are set for a response.
func (f Flags) IsResponse() bool {
	return f&FlagQR == FlagQR
}

// IsConflict returns true if the conflict flag is set.
func (f Flags) IsConflict() bool {
	return f&FlagC != 0
}

// IsTruncation returns true if the truncation flag is set.
func (f Flags) IsTruncation() bool {
	return f&FlagTC != 0
}

// IsTentative returns true if the tentative flag is set.
func (f Flags) IsTentative() bool {
	return f&FlagT != 0
}

// Opcode returns the 4-bit Opcode field (bits 1-4) as a value in the range 0-15.
func (f Flags) Opcode() uint8 {
	return uint8((f & MaskOpcode) >> ShiftOpcode)
}

// SetOpcode sets the 4-bit Opcode field (bits 1-4). Only the low 4 bits of
// opcode are used; higher bits are ignored.
func (f *Flags) SetOpcode(opcode uint8) {
	*f = (*f &^ MaskOpcode) | ((Flags(opcode) << ShiftOpcode) & MaskOpcode)
}

// RCODE returns the 4-bit RCODE field (bits 12-15) as a value in the range 0-15.
func (f Flags) RCODE() uint8 {
	return uint8(f & MaskRCODE)
}

// SetRCODE sets the 4-bit RCODE field (bits 12-15). Only the low 4 bits of
// rcode are used; higher bits are ignored.
func (f *Flags) SetRCODE(rcode uint8) {
	*f = (*f &^ MaskRCODE) | (Flags(rcode) & MaskRCODE)
}

// Z returns the 4-bit Z reserved field (bits 8-11) as a value in the range 0-15.
// Per RFC 4795 §2.1.1 these bits MUST be zero in conformant queries and responses.
func (f Flags) Z() uint8 {
	return uint8((f & MaskZ) >> ShiftZ)
}

// SetZ sets the 4-bit Z reserved field (bits 8-11). Only the low 4 bits of z
// are used; higher bits are ignored. Conformant implementations should leave
// this at zero.
func (f *Flags) SetZ(z uint8) {
	*f = (*f &^ MaskZ) | ((Flags(z) << ShiftZ) & MaskZ)
}

// Marshal encodes the Flags into a 2-byte big-endian representation.
func (f *Flags) Marshal() ([]byte, error) {
	marshalledData := make([]byte, 2)
	binary.BigEndian.PutUint16(marshalledData, uint16(*f))
	return marshalledData, nil
}

// Unmarshal decodes a 2-byte big-endian representation into the Flags receiver.
// It returns an error if the input slice is not exactly 2 bytes.
func (f *Flags) Unmarshal(data []byte) (int, error) {
	if len(data) != 2 {
		return 0, fmt.Errorf("invalid length: got %d bytes, want 2 bytes", len(data))
	}

	bytesRead := 0
	*f = Flags(binary.BigEndian.Uint16(data[0:2]))
	bytesRead += 2

	return bytesRead, nil
}

// String returns a string representation of the flags.
//
// The QR label is emitted only when the QR bit is set (response), matching the
// convention used in DNS trace output. Query messages (QR=0) produce no QR
// label. The Opcode and RCODE fields are 4-bit values and are rendered as
// "OPCODE=n"/"RCODE=n" only when non-zero, so the common case (standard query,
// no error) stays terse.
func (f Flags) String() string {
	flags := []string{}
	if f.IsResponse() {
		flags = append(flags, "QR")
	}
	if op := f.Opcode(); op != 0 {
		flags = append(flags, fmt.Sprintf("OPCODE=%d", op))
	}
	if f.IsConflict() {
		flags = append(flags, "C")
	}
	if f.IsTruncation() {
		flags = append(flags, "TC")
	}
	if f.IsTentative() {
		flags = append(flags, "T")
	}
	if rc := f.RCODE(); rc != 0 {
		flags = append(flags, fmt.Sprintf("RCODE=%d", rc))
	}
	return strings.Join(flags, "|")
}
