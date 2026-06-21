package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// DRS_EXTENSIONS is the variable-length capability blob exchanged at IDL_DRSBind
// ([MS-DRSR] 5.38): a byte count followed by Cb opaque bytes. The bytes carry a
// DRS_EXTENSIONS_INT (5.39) — see DRS_EXTENSIONS_INT and BuildClientExtensions.
//
// Rgb is an INLINE conformant array (`[size_is(cb)] BYTE rgb[]`), not a pointer to one:
// the maximum_count is hoisted ahead of Cb and no referent id is emitted. (idlgen
// defaulted this to `unique`, which would emit a stray referent id and fault the bind.)
type DRS_EXTENSIONS struct {
	Cb  ndr.DWORD
	Rgb []uint8 `ndr:"conformant,size_is=Cb"`
}
