// Package rpcinterface_22e5386d8b124bf0b0ec6a1ea419e366_1_0 is the descriptor for the NetEventForwarder RPC interface, abstract
// syntax 22e5386d-8b12-4bf0-b0ec-6a1ea419e366 version 1.0 ([MS-LREC]).
//
// The NetEventForwarder interface implements the Live Remote Event Capture (LREC)
// Protocol: a management station opens a context handle to a running event session on a
// target host, drains buffered events, and closes the session.
package rpcinterface_22e5386d8b124bf0b0ec6a1ea419e366_1_0

// IDL source: [MS-LREC] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-lrec/38d3f038-84ba-49ea-8828-3249ff2b5f9a
// A fetched copy is kept at ms-lrec.idl in the interface directory.

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// PipeName is retained for descriptor uniformity, but NetEventForwarder is NOT a
// named-pipe interface: per [MS-LREC] 2.1 the protocol is carried over ncacn_ip_tcp on a
// dynamic endpoint that the client resolves through the endpoint mapper (TCP/135). The
// client MUST authenticate with RPC_C_AUTHN_GSS_NEGOTIATE and SHOULD request
// RPC_C_AUTHN_LEVEL_PKT_PRIVACY. This constant is not the standard binding for this
// interface.
const PipeName = `\NetEventForwarder`

// Opnums for the on-the-wire methods.
const (
	OpnumRpcNetEventOpenSession  uint16 = 0
	OpnumRpcNetEventReceiveData  uint16 = 1
	OpnumRpcNetEventCloseSession uint16 = 2
)

// Status codes returned by this interface. Every method that returns a DWORD returns a
// Win32 error code ([MS-ERREF] 2.2): ERROR_SUCCESS on success, or a nonzero code on
// failure. [MS-LREC] 3.1.4.2 enumerates no specific failure codes — "all error values
// MUST be treated the same" — so only success is named; StatusString renders any other
// code as hex.
const (
	StatusSuccess uint32 = 0x00000000 // ERROR_SUCCESS
)

// SyntaxID returns the NetEventForwarder abstract syntax identifier:
// 22e5386d-8b12-4bf0-b0ec-6a1ea419e366, version 1.0.
func SyntaxID() syntax.SyntaxID {
	return syntax.SyntaxID{
		UUID:         guid.GUID{A: 0x22e5386d, B: 0x8b12, C: 0x4bf0, D: 0xb0ec, E: 0x6a1ea419e366},
		MajorVersion: 1,
		MinorVersion: 0,
	}
}

// StatusString returns a mnemonic for the documented status codes, otherwise the
// hex value.
func StatusString(status uint32) string {
	switch status {
	case StatusSuccess:
		return "ERROR_SUCCESS"
	default:
		return fmt.Sprintf("0x%08x", status)
	}
}

// OpnumToName maps each on-the-wire opnum to its method name; the single source of
// truth.
var OpnumToName = map[uint16]string{
	OpnumRpcNetEventOpenSession:  "RpcNetEventOpenSession",
	OpnumRpcNetEventReceiveData:  "RpcNetEventReceiveData",
	OpnumRpcNetEventCloseSession: "RpcNetEventCloseSession",
}

// NameToOpnum is the reverse of OpnumToName, built at init so the two never drift.
var NameToOpnum = func() map[string]uint16 {
	m := make(map[string]uint16, len(OpnumToName))
	for op, name := range OpnumToName {
		m[name] = op
	}
	return m
}()
