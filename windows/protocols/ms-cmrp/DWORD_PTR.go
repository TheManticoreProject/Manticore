package mscmrp

import "github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"

// DWORD_PTR is the [MS-DTYP] 2.2.9 pointer-precision unsigned type (unsigned
// __int3264). Under the classic DCE/RPC (NDR) transfer syntax used by this
// interface it is transmitted as a 4-byte value ([C706]); model it as a 32-bit
// word so the wire size is fixed and independent of the host word size (the
// 32-bit CI leg would otherwise diverge from the 64-bit one). Used by
// NOTIFICATION_RPC.DwNotifyKey ([MS-CMRP] 2.2.3.9).
type DWORD_PTR = ndr.DWORD
