// Package rpcinterface_2f5f6520ca461067b31900dd010662da_1_0 is the descriptor for the
// tapsrv (Telephony Server) RPC interface, abstract syntax
// 2f5f6520-ca46-1067-b319-00dd010662da version 1.0 ([MS-TRP]).
//
// An RPC interface is identified by its UUID and version, never by the named pipe it is
// reached over: the directory is named after the UUID with the version in the nested
// <maj>.<min>/ directory.
//
// This package holds only the interface-level descriptor (abstract syntax, transport
// endpoint, opnums, opnum<->name maps, status constants). NDR types live in the
// windows/protocols/ms-trp package (imported as mstrp) and method stubs in functions;
// both depend on this package, never the reverse.
package rpcinterface_2f5f6520ca461067b31900dd010662da_1_0

// IDL source: [MS-TRP] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-trp/e86aca98-76e9-4515-9de1-2cadb9084a2b
// A fetched copy is kept at ms-trp.idl in the interface directory.

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// PipeName is the IPC$-relative named pipe for the tapsrv interface. Per [MS-TRP] 2.1
// the client reaches the telephony server at the well-known endpoint \pipe\tapsrv
// (protocol sequence ncacn_np).
const PipeName = `\tapsrv`

// Opnums for the on-the-wire methods.
const (
	OpnumClientAttach  uint16 = 0
	OpnumClientRequest uint16 = 1
	OpnumClientDetach  uint16 = 2
)

// Status codes for this interface. ClientAttach returns 0 on success and a nonzero error
// code otherwise ([MS-TRP] 3.2.4.1); the void methods (ClientRequest, ClientDetach) carry
// their result inside the packed TAPI buffer rather than as an RPC return value. [MS-TRP]
// does not enumerate a named error-code table for the RPC return, so only success is
// modeled; StatusString renders any other value as hex.
const (
	StatusSuccess uint32 = 0x00000000 // success (return value 0)
)

// SyntaxID returns the tapsrv abstract syntax identifier:
// 2f5f6520-ca46-1067-b319-00dd010662da, version 1.0.
func SyntaxID() syntax.SyntaxID {
	return syntax.SyntaxID{
		UUID:         guid.GUID{A: 0x2f5f6520, B: 0xca46, C: 0x1067, D: 0xb319, E: 0x00dd010662da},
		MajorVersion: 1,
		MinorVersion: 0,
	}
}

// StatusString returns a mnemonic for the documented status codes, otherwise the
// hex value.
func StatusString(status uint32) string {
	switch status {
	case StatusSuccess:
		return "STATUS_SUCCESS"
	default:
		return fmt.Sprintf("0x%08x", status)
	}
}

// OpnumToName maps each on-the-wire opnum to its method name; the single source of
// truth.
var OpnumToName = map[uint16]string{
	OpnumClientAttach:  "ClientAttach",
	OpnumClientRequest: "ClientRequest",
	OpnumClientDetach:  "ClientDetach",
}

// NameToOpnum is the reverse of OpnumToName, built at init so the two never drift.
var NameToOpnum = func() map[string]uint16 {
	m := make(map[string]uint16, len(OpnumToName))
	for op, name := range OpnumToName {
		m[name] = op
	}
	return m
}()
