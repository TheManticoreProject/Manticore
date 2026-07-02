// Package dtyp provides the [MS-DTYP] common data types that recur across DCE/RPC
// interfaces (lsarpc, samr, srvsvc, …) as NDR-marshallable Go types, so each interface
// reuses one definition instead of redeclaring the same NDR-tagged structs.
//
// The types are driven by the declarative reflection walker in
// network/dcerpc/ndr (struct tags, the Marshaler escape hatch where needed), so they
// compose into request/response structs exactly like an interface's own structures. The
// package sits beside ndr and syntax rather than inside ndr: these are data types
// defined by a separate specification ([MS-DTYP], not the NDR transfer syntax), and
// they are version-neutral, used by both the connectionless (v4) and connection-
// oriented (v5) interfaces.
//
// References:
//   - [MS-DTYP] Windows Data Types:
//     https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dtyp/cca27429-5689-4a16-b2b4-9325d93e4ba2
//   - [MS-DTYP] Appendix A: Full IDL:
//     https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dtyp/24637f2d-238b-4d22-b44d-fe54b024280c
//   - [C706] DCE 1.1: RPC, Chapter 14 "Transfer Syntax NDR":
//     https://pubs.opengroup.org/onlinepubs/9629399/chap14.htm
package dtyp

// LARGE_INTEGER is the [MS-DTYP] 2.3.5 signed 64-bit integer. It is a named type rather
// than a bare int64 so declarations read like the IDL and it carries 8-octet NDR
// alignment ([C706] section 14.2.2).
//
// Reference: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dtyp/e904b1ba-f774-4203-ba1b-66485165ab1a
type LARGE_INTEGER int64

// ULARGE_INTEGER is the [MS-DTYP] 2.3.13 unsigned 64-bit integer.
//
// Reference: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dtyp/d7e6e5a5-6c77-4ae6-9bd5-3892b3c4641e
type ULARGE_INTEGER uint64

// FILETIME is the [MS-DTYP] 2.3.3 64-bit timestamp (100-nanosecond intervals since
// 1601-01-01 UTC) transmitted as two 32-bit halves, low then high. Unlike a bare uint64
// it carries 4-octet NDR alignment because both members are 32-bit ([C706] section
// 14.2.2), matching how Windows lays it out on the wire.
//
// Reference: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dtyp/2c57429b-fdd4-488f-b5fc-9e4cf020fcdf
type FILETIME struct {
	DwLowDateTime  uint32
	DwHighDateTime uint32
}

// Uint64 returns the FILETIME as a single 64-bit value (high half in the upper 32 bits).
func (f FILETIME) Uint64() uint64 {
	return uint64(f.DwHighDateTime)<<32 | uint64(f.DwLowDateTime)
}

// SYSTEMTIME is the [MS-DTYP] 2.3.13 wall-clock date and time, represented as eight
// consecutive 16-bit fields (year, month, day-of-week, day, hour, minute, second,
// millisecond). Each member is a 16-bit integer, so the structure carries 2-octet NDR
// alignment ([C706] section 14.2.2), matching how Windows lays it out on the wire.
//
// Reference: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dtyp/2fefe8dd-ab48-4e33-a7d5-7171455a9289
type SYSTEMTIME struct {
	WYear         uint16
	WMonth        uint16
	WDayOfWeek    uint16
	WDay          uint16
	WHour         uint16
	WMinute       uint16
	WSecond       uint16
	WMilliseconds uint16
}

// LUID is the [MS-DTYP] 2.3.7 locally unique identifier: a 64-bit value split into a
// low (unsigned) and high (signed) 32-bit part, transmitted in that order. It is used
// pervasively for privilege identifiers (for example by LsarLookupPrivilegeValue).
//
// Reference: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dtyp/48cbee2a-0790-45f2-8269-931d7083b2c3
type LUID struct {
	LowPart  uint32
	HighPart int32
}

// Uint64 returns the LUID as a single 64-bit value (HighPart in the upper 32 bits).
func (l LUID) Uint64() uint64 {
	return uint64(uint32(l.HighPart))<<32 | uint64(l.LowPart)
}

// LUIDFromUint64 splits a 64-bit value into a LUID.
func LUIDFromUint64(v uint64) LUID {
	return LUID{LowPart: uint32(v), HighPart: int32(v >> 32)}
}
