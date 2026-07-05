package msrrp

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// RPC_SECURITY_DESCRIPTOR ([MS-RRP] 2.2.8) carries a self-relative security descriptor
// for the key-security opnums (BaseRegGetKeySecurity / BaseRegSetKeySecurity).
//
// LpSecurityDescriptor is a [unique] pointer to a conformant-varying byte array whose
// maximum_count is CbInSecurityDescriptor (the buffer capacity) and whose actual_count is
// CbOutSecurityDescriptor (the number of valid bytes). The two are independent: the caller
// sets both fields explicitly, and only CbOutSecurityDescriptor bytes are transmitted.
//
//   - Set: point LpSecurityDescriptor at the marshalled descriptor and set both
//     CbInSecurityDescriptor and CbOutSecurityDescriptor to its length.
//   - Get: supply a buffer of capacity CbInSecurityDescriptor with CbOutSecurityDescriptor
//     = 0 (no valid bytes yet); the server fills the [out] descriptor.
type RPC_SECURITY_DESCRIPTOR struct {
	LpSecurityDescriptor    []uint8 `ndr:"unique,size_is=CbInSecurityDescriptor,varying,length_is=CbOutSecurityDescriptor"`
	CbInSecurityDescriptor  ndr.DWORD
	CbOutSecurityDescriptor ndr.DWORD
}
