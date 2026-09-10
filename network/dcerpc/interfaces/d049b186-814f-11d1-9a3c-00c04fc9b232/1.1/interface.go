// Package rpcinterface_d049b186814f11d19a3c00c04fc9b232_1_1 is the descriptor for the NtFrsApi RPC interface, abstract
// syntax d049b186-814f-11d1-9a3c-00c04fc9b232 version 1.1 ([MS-FRS1]).
//
// This package holds only the interface-level descriptor (abstract syntax,
// transport endpoint, opnums, and opnum<->name maps). The NDR wire types live in
// windows/protocols/ms-frs1 and the method stubs in functions; both depend on this
// package, never the reverse.
package rpcinterface_d049b186814f11d19a3c00c04fc9b232_1_1

// IDL source: [MS-FRS1] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-frs1/dd60a0d9-176a-46f4-9904-000172041b92
// A fetched copy is kept at ms-frs1.idl in the interface directory.

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// PipeName is empty: FRS has no named-pipe endpoint. Both FRS interfaces use only the
// ncacn_ip_tcp protocol sequence over a dynamic endpoint assigned by the RPC endpoint
// mapper (RPCSS, port 135), optionally pinned to a static TCP port ([MS-FRS1] 2.1).
const PipeName = ``

// Opnums for the on-the-wire methods. Opnums 0, 1, 2, 3, 6 are "not used on the wire"
// and are omitted.
const (
	OpnumNtFrsApi_Rpc_Set_DsPollingIntervalW uint16 = 4
	OpnumNtFrsApi_Rpc_Get_DsPollingIntervalW uint16 = 5
	OpnumNtFrsApi_Rpc_InfoW                  uint16 = 7
	OpnumNtFrsApi_Rpc_IsPathReplicated       uint16 = 8
	OpnumNtFrsApi_Rpc_WriterCommand          uint16 = 9
	OpnumNtFrsApi_Rpc_ForceReplication       uint16 = 10
)

// The status codes this interface returns are not declared here. FRSAPI methods return a
// Win32 error code ([MS-ERREF] 2.2): 0 on success, and all nonzero values are equivalent
// failures unless otherwise specified ([MS-FRS1] 3.2.4); a failed access check yields
// ERROR_ACCESS_DENIED. The whole of [MS-ERREF] 2.2 lives in
// github.com/TheManticoreProject/Manticore/windows/errors/win32 as the WIN32_ERROR type,
// and a subset repeated here would cover a fraction of that 2703-code table while
// drifting from it. Convert a returned status with win32.WIN32_ERROR(status) and compare
// against win32.ERROR_SUCCESS and the rest.
//
// Both codes this descriptor used to declare have an [MS-ERREF] 2.2 row under the name
// the specification uses, so neither stays local: StatusSuccess is win32.ERROR_SUCCESS
// (0x00000000) and ErrorAccessDenied is win32.ERROR_ACCESS_DENIED (0x00000005).
// StatusString went with them.
//
// The NtFrs service errors an FRSAPI method reports are the FRS_ERR_* family, which the
// shared table carries at 0x00001F41..0x00001F51. They are unrelated to the FRS_ERROR_*
// family [MS-FRS2] defines at 0x23xx, which [MS-ERREF] 2.2 has no rows for at all.

// SyntaxID returns the NtFrsApi abstract syntax identifier:
// d049b186-814f-11d1-9a3c-00c04fc9b232, version 1.1.
func SyntaxID() syntax.SyntaxID {
	return syntax.SyntaxID{
		UUID:         guid.GUID{A: 0xd049b186, B: 0x814f, C: 0x11d1, D: 0x9a3c, E: 0x00c04fc9b232},
		MajorVersion: 1,
		MinorVersion: 1,
	}
}

// OpnumToName maps each on-the-wire opnum to its method name; the single source of
// truth.
var OpnumToName = map[uint16]string{
	OpnumNtFrsApi_Rpc_Set_DsPollingIntervalW: "NtFrsApi_Rpc_Set_DsPollingIntervalW",
	OpnumNtFrsApi_Rpc_Get_DsPollingIntervalW: "NtFrsApi_Rpc_Get_DsPollingIntervalW",
	OpnumNtFrsApi_Rpc_InfoW:                  "NtFrsApi_Rpc_InfoW",
	OpnumNtFrsApi_Rpc_IsPathReplicated:       "NtFrsApi_Rpc_IsPathReplicated",
	OpnumNtFrsApi_Rpc_WriterCommand:          "NtFrsApi_Rpc_WriterCommand",
	OpnumNtFrsApi_Rpc_ForceReplication:       "NtFrsApi_Rpc_ForceReplication",
}

// NameToOpnum is the reverse of OpnumToName, built at init so the two never drift.
var NameToOpnum = func() map[string]uint16 {
	m := make(map[string]uint16, len(OpnumToName))
	for op, name := range OpnumToName {
		m[name] = op
	}
	return m
}()
