// Package rpcinterface_d049b186814f11d19a3c00c04fc9b232_1_1 is the descriptor for the NtFrsApi RPC interface, abstract
// syntax d049b186-814f-11d1-9a3c-00c04fc9b232 version 1.1 ([MS-FRS1]).
//
// This package holds only the interface-level descriptor (abstract syntax,
// transport endpoint, opnums, opnum<->name maps, and status constants). The NDR
// wire types live in windows/protocols/ms-frs1 and the method stubs in functions;
// both depend on this package, never the reverse.
package rpcinterface_d049b186814f11d19a3c00c04fc9b232_1_1

import (
	"fmt"

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

// Status codes returned by this interface. FRSAPI methods return a Win32 error code
// ([MS-ERREF] 2.2): 0 on success, and all nonzero values are equivalent failures unless
// otherwise specified ([MS-FRS1] 3.2.4). Failed access checks yield ERROR_ACCESS_DENIED.
const (
	StatusSuccess     uint32 = 0x00000000 // ERROR_SUCCESS
	ErrorAccessDenied uint32 = 0x00000005 // ERROR_ACCESS_DENIED
)

// SyntaxID returns the NtFrsApi abstract syntax identifier:
// d049b186-814f-11d1-9a3c-00c04fc9b232, version 1.1.
func SyntaxID() syntax.SyntaxID {
	return syntax.SyntaxID{
		UUID:         guid.GUID{A: 0xd049b186, B: 0x814f, C: 0x11d1, D: 0x9a3c, E: 0x00c04fc9b232},
		MajorVersion: 1,
		MinorVersion: 1,
	}
}

// StatusString returns a mnemonic for the documented status codes, otherwise the
// hex value.
func StatusString(status uint32) string {
	switch status {
	case StatusSuccess:
		return "ERROR_SUCCESS"
	case ErrorAccessDenied:
		return "ERROR_ACCESS_DENIED"
	default:
		return fmt.Sprintf("0x%08x", status)
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
