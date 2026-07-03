// Package rpcinterface_ae33069ba2a846eea235ddfd339be281_1_0 is the descriptor for the IRPCRemoteObject RPC interface, abstract
// syntax ae33069b-a2a8-46ee-a235-ddfd339be281 version 1.0 ([MS-PAN]).
//
// IRPCRemoteObject creates and deletes the remote objects (context handles) that
// IRPCAsyncNotify calls take as their registration/channel arguments ([MS-PAN] 3.1.2).
package rpcinterface_ae33069ba2a846eea235ddfd339be281_1_0

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// PipeName is empty: IRPCRemoteObject has no named-pipe endpoint. Per [MS-PAN] section
// 2.1 the interface is bound over RPC-on-TCP (ncacn_ip_tcp) at an RPC dynamic endpoint
// assigned by the endpoint mapper ([C706] Part 4), located by the interface UUID rather
// than a well-known port or pipe.
const PipeName = ``

// Opnums for the on-the-wire methods.
const (
	OpnumIRPCRemoteObject_Create uint16 = 0
	OpnumIRPCRemoteObject_Delete uint16 = 1
)

// Status codes. IRPCRemoteObject_Create returns an HRESULT: ZERO (S_OK, 0x00000000) on
// success, or a common [MS-ERREF] HRESULT on failure ([MS-PAN] 3.1.2.4).
// IRPCRemoteObject_Delete returns void. ([MS-ERREF] section 2.1.1.)
const (
	StatusSuccess uint32 = 0x00000000 // S_OK

	// Common [MS-ERREF] HRESULTs a server may return.
	ErrorAccessDenied uint32 = 0x80070005 // E_ACCESSDENIED
	ErrorOutOfMemory  uint32 = 0x8007000E // E_OUTOFMEMORY
	ErrorInvalidArg   uint32 = 0x80070057 // E_INVALIDARG
)

// SyntaxID returns the IRPCRemoteObject abstract syntax identifier:
// ae33069b-a2a8-46ee-a235-ddfd339be281, version 1.0.
func SyntaxID() syntax.SyntaxID {
	return syntax.SyntaxID{
		UUID:         guid.GUID{A: 0xae33069b, B: 0xa2a8, C: 0x46ee, D: 0xa235, E: 0xddfd339be281},
		MajorVersion: 1,
		MinorVersion: 0,
	}
}

// StatusString returns a mnemonic for the documented status codes, otherwise the
// hex value.
func StatusString(status uint32) string {
	switch status {
	case StatusSuccess:
		return "S_OK"
	case ErrorAccessDenied:
		return "E_ACCESSDENIED"
	case ErrorOutOfMemory:
		return "E_OUTOFMEMORY"
	case ErrorInvalidArg:
		return "E_INVALIDARG"
	default:
		return fmt.Sprintf("0x%08x", status)
	}
}

// OpnumToName maps each on-the-wire opnum to its method name; the single source of
// truth.
var OpnumToName = map[uint16]string{
	OpnumIRPCRemoteObject_Create: "IRPCRemoteObject_Create",
	OpnumIRPCRemoteObject_Delete: "IRPCRemoteObject_Delete",
}

// NameToOpnum is the reverse of OpnumToName, built at init so the two never drift.
var NameToOpnum = func() map[string]uint16 {
	m := make(map[string]uint16, len(OpnumToName))
	for op, name := range OpnumToName {
		m[name] = op
	}
	return m
}()
