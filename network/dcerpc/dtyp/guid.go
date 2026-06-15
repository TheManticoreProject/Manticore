package dtyp

import (
	"encoding/binary"

	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// GUID is the [MS-DTYP] 2.3.4.2 GUID in its NDR-marshallable form: Data1 (4 octets),
// Data2 (2), Data3 (2), and Data4 (8 opaque octets), for 16 octets total with 4-octet
// alignment. Data1/2/3 are little-endian integers; Data4 is transmitted verbatim.
//
// The reflection walker cannot use windows/guid.GUID directly: its Go layout ends in a
// uint64, which over-aligns the struct to 8 and inserts interior padding, so it would
// marshal as 24 octets instead of 16. This type mirrors the wire layout exactly. Use
// NewGUID / GUID to convert to and from windows/guid.GUID.
//
// Reference: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dtyp/49e490b8-f972-45d6-a3a4-99f924998d97
type GUID struct {
	Data1 uint32
	Data2 uint16
	Data3 uint16
	Data4 [8]byte
}

// NewGUID converts a windows/guid.GUID to its NDR GUID form.
func NewGUID(g guid.GUID) GUID {
	b := g.ToBytes() // 16 octets: Data1/2/3 little-endian, Data4 verbatim
	var x GUID
	x.Data1 = binary.LittleEndian.Uint32(b[0:4])
	x.Data2 = binary.LittleEndian.Uint16(b[4:6])
	x.Data3 = binary.LittleEndian.Uint16(b[6:8])
	copy(x.Data4[:], b[8:16])
	return x
}

// GUID converts back to a windows/guid.GUID.
func (g GUID) GUID() guid.GUID {
	var b [16]byte
	binary.LittleEndian.PutUint32(b[0:4], g.Data1)
	binary.LittleEndian.PutUint16(b[4:6], g.Data2)
	binary.LittleEndian.PutUint16(b[6:8], g.Data3)
	copy(b[8:16], g.Data4[:])
	var out guid.GUID
	out.FromRawBytes(b[:])
	return out
}

// String renders the GUID in the standard "D" format.
func (g GUID) String() string {
	w := g.GUID()
	return w.ToFormatD()
}
