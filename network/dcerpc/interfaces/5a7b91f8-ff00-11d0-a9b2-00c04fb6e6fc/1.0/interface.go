// Package rpcinterface_5a7b91f8ff0011d0a9b200c04fb6e6fc_1_0 is the descriptor for the msgsvcsend RPC interface, abstract
// syntax 5a7b91f8-ff00-11d0-a9b2-00c04fb6e6fc version 1.0 ([MS-MSRP]).
//
// The PipeName and the doc comments are not derivable from the IDL and were
// filled in by hand from [MS-MSRP].
package rpcinterface_5a7b91f8ff0011d0a9b200c04fb6e6fc_1_0

// IDL source: [MS-MSRP] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-msrp/181965ff-fab4-4ad4-a8d7-16b444cc4e66
// A fetched copy is kept at ms-msrp.idl in the interface directory.

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// PipeName is the IPC$-relative named pipe for the msgsvcsend interface. The primary
// transport for NetrSendMessage is RPC over UDP (ncadg_ip_udp) with dynamic endpoints;
// when RPC over Named Pipes (ncacn_np) is used, [MS-MSRP] 2.1 mandates \PIPE\MSGSVC.
const PipeName = `\msgsvc`

// Opnums for the on-the-wire methods ([MS-MSRP] 3.2.4).
const (
	OpnumNetrSendMessage uint16 = 0
)

// The NET_API_STATUS / error_status_t codes NetrSendMessage returns ([MS-MSRP] 3.2.4.1,
// [MS-ERREF] 2.2, lmerr.h) are not declared here. The whole of [MS-ERREF] 2.2, the
// NERR_* range included, lives in
// github.com/TheManticoreProject/Manticore/windows/errors/win32 as the WIN32_ERROR type,
// and a subset repeated here would cover a fraction of that 2703-code table while drifting
// from it. Convert a returned status with win32.WIN32_ERROR(status) and compare against
// win32.NERR_Success and the rest.

// SyntaxID returns the msgsvcsend abstract syntax identifier:
// 5a7b91f8-ff00-11d0-a9b2-00c04fb6e6fc, version 1.0.
func SyntaxID() syntax.SyntaxID {
	return syntax.SyntaxID{
		UUID:         guid.GUID{A: 0x5a7b91f8, B: 0xff00, C: 0x11d0, D: 0xa9b2, E: 0x00c04fb6e6fc},
		MajorVersion: 1,
		MinorVersion: 0,
	}
}

// OpnumToName maps each on-the-wire opnum to its method name; the single source of
// truth.
var OpnumToName = map[uint16]string{
	OpnumNetrSendMessage: "NetrSendMessage",
}

// NameToOpnum is the reverse of OpnumToName, built at init so the two never drift.
var NameToOpnum = func() map[string]uint16 {
	m := make(map[string]uint16, len(OpnumToName))
	for op, name := range OpnumToName {
		m[name] = op
	}
	return m
}()
