package types

import (
	"github.com/TheManticoreProject/Manticore/windows/ms_dtyp/common/data_structures"
	"github.com/TheManticoreProject/Manticore/windows/ms_dtyp/common/data_types"
)

// Base MS-DTYP scalar aliases reused across the SMB2 wire structures. These
// mirror the SMB 1.0 passthrough types (network/smb/smb_v10/types) so that the
// two packages share a consistent vocabulary, with the addition of the 64-bit
// unsigned types that SMB2 relies on for its identifiers (MessageId, SessionId,
// AsyncId, file sizes, ...).
type UCHAR = data_types.UCHAR
type USHORT = data_types.USHORT
type ULONG = data_types.ULONG

type CHAR = data_types.CHAR
type SHORT = data_types.SHORT
type LONG = data_types.LONG

type DWORD = data_types.DWORD

// 64-bit unsigned aliases. SMB2 widens many SMB1 16/32-bit fields to 64 bits.
type UINT64 = data_types.UINT64
type ULONG64 = data_types.ULONG64

type LARGE_INTEGER = data_structures.LARGE_INTEGER

// FILETIME is the 8-byte 64-bit timestamp used throughout SMB2 (MS-DTYP 2.3.3).
type FILETIME = data_structures.FILETIME
