// Package rpcinterface_2f5f6520ca461067b31900dd010662da_1_0 is the descriptor for the
// tapsrv (Telephony Server) RPC interface, abstract syntax
// 2f5f6520-ca46-1067-b319-00dd010662da version 1.0 ([MS-TRP]).
//
// An RPC interface is identified by its UUID and version, never by the named pipe it is
// reached over: the directory is named after the UUID with the version in the nested
// <maj>.<min>/ directory.
//
// This package holds only the interface-level descriptor (abstract syntax, transport
// endpoint, opnums, opnum<->name maps). NDR types live in the windows/protocols/ms-trp
// package (imported as mstrp) and method stubs in functions; both depend on this package,
// never the reverse.
package rpcinterface_2f5f6520ca461067b31900dd010662da_1_0

// IDL source: [MS-TRP] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-trp/e86aca98-76e9-4515-9de1-2cadb9084a2b
// A fetched copy is kept at ms-trp.idl in the interface directory.

import (
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

// ClientAttach returns 0 on success and otherwise a nonzero error code "as specified in
// [MS-ERREF]" ([MS-TRP] 3.2.4.1); the void methods (ClientRequest, ClientDetach) carry
// their result inside the packed TAPI buffer rather than as an RPC return value. Those
// codes are not declared here. The whole of [MS-ERREF] 2.2 lives in
// github.com/TheManticoreProject/Manticore/windows/errors/win32 as the WIN32_ERROR type,
// and a subset repeated here would cover a fraction of that 2703-code table while
// drifting from it. Convert a returned status with win32.WIN32_ERROR(status) and compare
// against win32.ERROR_SUCCESS and the rest. The one success value this interface used to
// declare has a row in [MS-ERREF] 2.2, so nothing is kept local.
//
// Two values the specification mentions for this return sit outside that table and stay
// hexadecimal, exactly as they did before: LINEERR_OPERATIONFAILED (0x80000048), a TAPI
// code from the 0x8000xxxx block [MS-ERREF] 2.2 does not cover, and the -19 (0xFFFFFFED)
// that reports a client without administrator access. Neither has a row, so neither can
// be misnamed out of the Win32 table either.

// SyntaxID returns the tapsrv abstract syntax identifier:
// 2f5f6520-ca46-1067-b319-00dd010662da, version 1.0.
func SyntaxID() syntax.SyntaxID {
	return syntax.SyntaxID{
		UUID:         guid.GUID{A: 0x2f5f6520, B: 0xca46, C: 0x1067, D: 0xb319, E: 0x00dd010662da},
		MajorVersion: 1,
		MinorVersion: 0,
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
