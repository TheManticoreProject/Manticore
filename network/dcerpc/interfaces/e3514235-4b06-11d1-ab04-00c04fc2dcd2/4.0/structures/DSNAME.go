package structures

import (
	"unicode/utf16"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// DSNAME identifies a directory object by GUID, SID, and/or distinguished name
// ([MS-DRSR] 5.50). StringName is the IDL's inline conformant array
// `[size_is(NameLen+1)] WCHAR StringName[]` — NOT a pointer (idlgen defaulted it to
// `unique`, which would emit a stray referent id). As the trailing conformant member its
// maximum_count is hoisted ahead of StructLen; the codec derives that count from the
// slice length, so callers MUST size StringName to NameLen+1 (the DN plus its NUL
// terminator). Build one with NewDSNameFromGUID for GUID-based addressing (EXOP_REPL_OBJ).
type DSNAME struct {
	StructLen  ndr.DWORD
	SidLen     ndr.DWORD
	Guid       UUID
	Sid        NT4SID
	NameLen    ndr.DWORD
	StringName []uint16 `ndr:"conformant"`
}

// NewDSNameFromGUID builds a DSNAME that addresses an object solely by its objectGUID,
// with no SID and an empty (NUL-only) name — the form used to replicate a single object
// by GUID (IDL_DRSGetNCChanges EXOP_REPL_OBJ). NameLen is 0 and StringName is the single
// terminating NUL, so its maximum_count is 1. StructLen is the fixed prefix size (the
// fields through NameLen plus the one-WCHAR name): 4+4+16+28+4+2 = 58 bytes.
func NewDSNameFromGUID(g guid.GUID) DSNAME {
	return DSNAME{
		StructLen:  58,
		Guid:       UUIDFromGUID(g),
		NameLen:    0,
		StringName: []uint16{0},
	}
}

// NewDSNameFromDN builds a DSNAME that addresses an object by its distinguished name,
// with no GUID or SID — the form used to name a naming context head for full-NC
// replication (IDL_DRSGetNCChanges with no EXOP). NameLen is the DN's UTF-16 code-unit
// count; StringName is the DN plus its NUL terminator (so maximum_count is NameLen+1).
// StructLen is the fixed prefix (StructLen..NameLen = 56 octets) plus the name bytes.
func NewDSNameFromDN(dn string) DSNAME {
	units := utf16.Encode([]rune(dn))
	name := make([]uint16, len(units)+1)
	copy(name, units) // trailing element stays 0 (NUL terminator)
	return DSNAME{
		StructLen:  ndr.DWORD(56 + 2*len(name)),
		NameLen:    ndr.DWORD(len(units)),
		StringName: name,
	}
}
