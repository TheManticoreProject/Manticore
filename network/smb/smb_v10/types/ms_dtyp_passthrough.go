package types

import (
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

type UCHAR = msdtyp.UCHAR
type USHORT = msdtyp.USHORT
type ULONG = msdtyp.ULONG

type CHAR = msdtyp.CHAR
type SHORT = msdtyp.SHORT
type LONG = msdtyp.LONG

type DWORD = msdtyp.DWORD

// LARGE_INTEGER is SMB's fixed-layout 64-bit value, kept as a struct{QuadPart} local to
// the SMB types package. The shared msdtyp.LARGE_INTEGER is the NDR-oriented named scalar
// (int64); SMB marshals the QuadPart field directly into its little-endian wire buffers,
// so it keeps the struct form here.
type LARGE_INTEGER struct {
	QuadPart uint64
}

type FILETIME = msdtyp.FILETIME
