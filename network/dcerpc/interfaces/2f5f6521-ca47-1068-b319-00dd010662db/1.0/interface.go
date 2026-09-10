// Package rpcinterface_2f5f6521ca471068b31900dd010662db_1_0 is the descriptor for the
// remotesp (Remote Service Provider) RPC interface, abstract syntax
// 2f5f6521-ca47-1068-b319-00dd010662db version 1.0 ([MS-TRP]).
//
// remotesp is a reverse/callback interface: the telephony server acts as the RPC client
// and invokes these methods on the client, which hosts the remotesp server, to push
// asynchronous TAPI event notifications.
//
// An RPC interface is identified by its UUID and version, never by the transport it is
// reached over: the directory is named after the UUID with the version in the nested
// <maj>.<min>/ directory.
//
// This package holds only the interface-level descriptor (abstract syntax, transport
// endpoint, opnums, opnum<->name maps). NDR types live in the windows/protocols/ms-trp
// package (imported as mstrp) and method stubs in functions; both depend on this package,
// never the reverse.
package rpcinterface_2f5f6521ca471068b31900dd010662db_1_0

// IDL source: [MS-TRP] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-trp/e86aca98-76e9-4515-9de1-2cadb9084a2b
// A fetched copy is kept at ms-trp.idl in the interface directory.

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// PipeName is empty for the remotesp interface: [MS-TRP] 2.1 specifies that the client
// chooses the RPC protocol sequence and endpoint when it calls ClientAttach, and the
// telephony server establishes the reverse connection to that client-supplied endpoint.
// There is therefore no fixed well-known named pipe. It is retained, empty, for
// descriptor uniformity across interfaces.
const PipeName = ``

// Opnums for the on-the-wire methods.
const (
	OpnumRemoteSPAttach    uint16 = 0
	OpnumRemoteSPEventProc uint16 = 1
	OpnumRemoteSPDetach    uint16 = 2
)

// RemoteSPAttach returns 0 on success and otherwise a nonzero error code "as specified in
// [MS-ERREF]" ([MS-TRP] 3.1.4.1); the void methods (RemoteSPEventProc, RemoteSPDetach)
// carry no RPC return value. Those codes are not declared here. The whole of [MS-ERREF]
// 2.2 lives in github.com/TheManticoreProject/Manticore/windows/errors/win32 as the
// WIN32_ERROR type, and a subset repeated here would cover a fraction of that 2703-code
// table while drifting from it. Convert a returned status with win32.WIN32_ERROR(status)
// and compare against win32.ERROR_SUCCESS and the rest. The one success value this
// interface used to declare has a row in [MS-ERREF] 2.2, so nothing is kept local.
//
// The specification names no failure code for this return in particular, but the
// neighbouring tapsrv ClientAttach return does name LINEERR_OPERATIONFAILED (0x80000048),
// a TAPI code from the 0x8000xxxx block [MS-ERREF] 2.2 does not cover. Such a value has
// no row in the shared table, so it stays hexadecimal here exactly as it did before and
// cannot be misnamed out of the Win32 space.

// SyntaxID returns the remotesp abstract syntax identifier:
// 2f5f6521-ca47-1068-b319-00dd010662db, version 1.0.
func SyntaxID() syntax.SyntaxID {
	return syntax.SyntaxID{
		UUID:         guid.GUID{A: 0x2f5f6521, B: 0xca47, C: 0x1068, D: 0xb319, E: 0x00dd010662db},
		MajorVersion: 1,
		MinorVersion: 0,
	}
}

// OpnumToName maps each on-the-wire opnum to its method name; the single source of
// truth.
var OpnumToName = map[uint16]string{
	OpnumRemoteSPAttach:    "RemoteSPAttach",
	OpnumRemoteSPEventProc: "RemoteSPEventProc",
	OpnumRemoteSPDetach:    "RemoteSPDetach",
}

// NameToOpnum is the reverse of OpnumToName, built at init so the two never drift.
var NameToOpnum = func() map[string]uint16 {
	m := make(map[string]uint16, len(OpnumToName))
	for op, name := range OpnumToName {
		m[name] = op
	}
	return m
}()
