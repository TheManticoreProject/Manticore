// Package rpcinterface_ae33069ba2a846eea235ddfd339be281_1_0 is the descriptor for the IRPCRemoteObject RPC interface, abstract
// syntax ae33069b-a2a8-46ee-a235-ddfd339be281 version 1.0 ([MS-PAN]).
//
// IRPCRemoteObject creates and deletes the remote objects (context handles) that
// IRPCAsyncNotify calls take as their registration/channel arguments ([MS-PAN] 3.1.2).
package rpcinterface_ae33069ba2a846eea235ddfd339be281_1_0

// IDL source: [MS-PAN] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-pan/3161e1b8-098f-4f42-8a58-7e342114b643
// A fetched copy is kept at ms-pan.idl in the interface directory.

import (
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

// The HRESULTs IRPCRemoteObject_Create returns ([MS-PAN] 3.1.2.4) are not declared here.
// The whole of [MS-ERREF] 2.1.1 lives in
// github.com/TheManticoreProject/Manticore/windows/errors/hresult as the HRESULT type,
// and the four values repeated here covered four of that 2928-value table while drifting
// from it — [MS-PAN] 3.1.2.4 names no closed set, only "a common [MS-ERREF] HRESULT", so
// the subset could name none of the rest. Convert a returned status with
// hresult.HRESULT(status) and compare against hresult.S_OK;
// hresult.HRESULT.IsSuccess covers the whole success range, since for an HRESULT success
// is a severity bit rather than the single value zero. IRPCRemoteObject_Delete returns
// void and carries no status at all.

// SyntaxID returns the IRPCRemoteObject abstract syntax identifier:
// ae33069b-a2a8-46ee-a235-ddfd339be281, version 1.0.
func SyntaxID() syntax.SyntaxID {
	return syntax.SyntaxID{
		UUID:         guid.GUID{A: 0xae33069b, B: 0xa2a8, C: 0x46ee, D: 0xa235, E: 0xddfd339be281},
		MajorVersion: 1,
		MinorVersion: 0,
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
