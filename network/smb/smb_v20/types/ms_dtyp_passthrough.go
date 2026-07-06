package types

import (
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// Base MS-DTYP scalar aliases reused across the SMB2 wire structures. These
// mirror the SMB 1.0 passthrough types (network/smb/smb_v10/types) so that the
// two packages share a consistent vocabulary, with the addition of the 64-bit
// unsigned types that SMB2 relies on for its identifiers (MessageId, SessionId,
// AsyncId, file sizes, ...).
type UCHAR = msdtyp.UCHAR
type USHORT = msdtyp.USHORT
type ULONG = msdtyp.ULONG

type CHAR = msdtyp.CHAR
type SHORT = msdtyp.SHORT
type LONG = msdtyp.LONG

type DWORD = msdtyp.DWORD

// 64-bit unsigned aliases. SMB2 widens many SMB1 16/32-bit fields to 64 bits.
type UINT64 = msdtyp.UINT64
type ULONG64 = msdtyp.ULONG64

// LARGE_INTEGER is SMB's fixed-layout 64-bit value, kept as a struct{QuadPart} local to
// the SMB types package. The shared msdtyp.LARGE_INTEGER is the NDR-oriented named scalar
// (int64); SMB marshals the QuadPart field directly into its little-endian wire buffers,
// so it keeps the struct form here.
type LARGE_INTEGER struct {
	QuadPart uint64
}

// FILETIME is the 8-byte 64-bit timestamp used throughout SMB2 (MS-DTYP 2.3.3).
type FILETIME = msdtyp.FILETIME
