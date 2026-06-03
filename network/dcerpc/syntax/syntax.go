// Package syntax models DCE/RPC presentation syntax identifiers (p_syntax_id_t).
//
// A syntax identifier names either an abstract syntax (an interface UUID and
// version) or a transfer syntax (an encoding such as NDR). On the wire it is a
// 16-byte UUID followed by a 4-byte version: a 16-bit major version and a 16-bit
// minor version, both little-endian ([C706] section 12.6).
package syntax

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// Size is the marshalled size, in bytes, of a syntax identifier: 16-byte UUID plus
// a 2-byte major and 2-byte minor version.
const Size = 20

// SyntaxID identifies an abstract or transfer syntax.
type SyntaxID struct {
	UUID         guid.GUID
	MajorVersion uint16
	MinorVersion uint16
}

// NDRTransferSyntax returns the identifier for the NDR (Network Data Representation)
// transfer syntax, version 2.0: 8a885d04-1ceb-11c9-9fe8-08002b104860.
func NDRTransferSyntax() SyntaxID {
	return SyntaxID{
		UUID:         guid.GUID{A: 0x8a885d04, B: 0x1ceb, C: 0x11c9, D: 0x9fe8, E: 0x08002b104860},
		MajorVersion: 2,
		MinorVersion: 0,
	}
}

// NDR64TransferSyntax returns the identifier for the NDR64 transfer syntax, version
// 1.0: 71710533-beba-4937-8319-b5dbef9ccc36.
func NDR64TransferSyntax() SyntaxID {
	return SyntaxID{
		UUID:         guid.GUID{A: 0x71710533, B: 0xbeba, C: 0x4937, D: 0x8319, E: 0xb5dbef9ccc36},
		MajorVersion: 1,
		MinorVersion: 0,
	}
}

// Equal reports whether two syntax identifiers are identical.
func (s SyntaxID) Equal(other SyntaxID) bool {
	return s.UUID.Equal(&other.UUID) && s.MajorVersion == other.MajorVersion && s.MinorVersion == other.MinorVersion
}

// Marshal serializes the syntax identifier into its 20-byte wire form.
func (s SyntaxID) Marshal() ([]byte, error) {
	buf := make([]byte, Size)
	copy(buf[0:16], s.UUID.ToBytes())
	binary.LittleEndian.PutUint16(buf[16:18], s.MajorVersion)
	binary.LittleEndian.PutUint16(buf[18:20], s.MinorVersion)
	return buf, nil
}

// Unmarshal parses a syntax identifier from data and returns the number of bytes
// consumed.
func (s *SyntaxID) Unmarshal(data []byte) (int, error) {
	if len(data) < Size {
		return 0, fmt.Errorf("syntax identifier truncated: have %d bytes, need %d", len(data), Size)
	}
	s.UUID.FromRawBytes(data[0:16])
	s.MajorVersion = binary.LittleEndian.Uint16(data[16:18])
	s.MinorVersion = binary.LittleEndian.Uint16(data[18:20])
	return Size, nil
}

// String returns a human-readable representation, e.g. "8a885d04-...-08002b104860 v2.0".
func (s SyntaxID) String() string {
	return fmt.Sprintf("%s v%d.%d", s.UUID.ToFormatD(), s.MajorVersion, s.MinorVersion)
}
