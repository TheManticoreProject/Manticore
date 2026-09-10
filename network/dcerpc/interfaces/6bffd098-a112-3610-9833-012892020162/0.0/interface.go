// Package rpcinterface_6bffd098a11236109833012892020162_0_0 is the descriptor for the
// browser (\browser) RPC interface, abstract syntax
// 6bffd098-a112-3610-9833-012892020162 version 0.0 ([MS-BRWSA]).
//
// An RPC interface is identified by its UUID and version, never by the named pipe it is
// reached over. This package holds only the interface-level descriptor (abstract syntax,
// transport endpoint, opnums, opnum<->name maps, and status constants). NDR types live in
// windows/protocols/ms-brwsa and method stubs in functions; both depend on this package,
// never the reverse.
package rpcinterface_6bffd098a11236109833012892020162_0_0

// IDL source: [MS-BRWSA] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-brwsa/c20c5c21-d285-4e98-8480-36922da69adf
// A fetched copy is kept at ms-brwsa.idl in the interface directory.

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// PipeName is the IPC$-relative named pipe for the browser interface. The CIFS Browser
// Auxiliary Protocol is available only over the \PIPE\browser named pipe ([MS-BRWSA] 2.1).
const PipeName = `\browser`

// Opnums for the on-the-wire methods. Opnums 0, 1, 3, 4, 5, 6, 7, 8, 9, 10, 11 are "not used on the wire"
// and are omitted.
const (
	OpnumI_BrowserrQueryOtherDomains uint16 = 2
)

// The NET_API_STATUS codes I_BrowserrQueryOtherDomains returns ([MS-BRWSA] 3.1.4.1) are
// not declared here. A NET_API_STATUS is a Win32 error code, and the whole of
// [MS-ERREF] 2.2 lives in
// github.com/TheManticoreProject/Manticore/windows/errors/win32 as the WIN32_ERROR type,
// so a subset repeated here would cover a fraction of that 2703-code table while drifting
// from it. Convert a returned status with win32.WIN32_ERROR(status) and compare against
// win32.NERR_Success, which is the zero success the specification calls NERR_Success, and
// against win32.ERROR_MORE_DATA and the rest.

// SyntaxID returns the browser abstract syntax identifier:
// 6bffd098-a112-3610-9833-012892020162, version 0.0.
func SyntaxID() syntax.SyntaxID {
	return syntax.SyntaxID{
		UUID:         guid.GUID{A: 0x6bffd098, B: 0xa112, C: 0x3610, D: 0x9833, E: 0x012892020162},
		MajorVersion: 0,
		MinorVersion: 0,
	}
}

// OpnumToName maps each on-the-wire opnum to its method name; the single source of
// truth.
var OpnumToName = map[uint16]string{
	OpnumI_BrowserrQueryOtherDomains: "I_BrowserrQueryOtherDomains",
}

// NameToOpnum is the reverse of OpnumToName, built at init so the two never drift.
var NameToOpnum = func() map[string]uint16 {
	m := make(map[string]uint16, len(OpnumToName))
	for op, name := range OpnumToName {
		m[name] = op
	}
	return m
}()
