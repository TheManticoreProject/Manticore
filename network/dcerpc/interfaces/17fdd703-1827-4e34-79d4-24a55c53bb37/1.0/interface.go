// Package rpcinterface_17fdd70318274e3479d424a55c53bb37_1_0 is the descriptor for the msgsvc RPC interface, abstract
// syntax 17fdd703-1827-4e34-79d4-24a55c53bb37 version 1.0 ([MS-MSRP]).
//
// The PipeName and the doc comments are not derivable from the IDL and were
// filled in by hand from [MS-MSRP].
package rpcinterface_17fdd70318274e3479d424a55c53bb37_1_0

// IDL source: [MS-MSRP] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-msrp/181965ff-fab4-4ad4-a8d7-16b444cc4e66
// A fetched copy is kept at ms-msrp.idl in the interface directory.

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// PipeName is the IPC$-relative named pipe for the msgsvc interface. The message-name
// management methods use RPC over Named Pipes (ncacn_np); [MS-MSRP] 2.1 mandates the
// pipe \PIPE\MSGSVC.
const PipeName = `\msgsvc`

// Opnums for the on-the-wire methods ([MS-MSRP] 3.1.4).
const (
	OpnumNetrMessageNameAdd     uint16 = 0
	OpnumNetrMessageNameEnum    uint16 = 1
	OpnumNetrMessageNameGetInfo uint16 = 2
	OpnumNetrMessageNameDel     uint16 = 3
)

// The NET_API_STATUS codes msgsvc methods return ([MS-MSRP] 3.1.4, [MS-ERREF] 2.2,
// lmerr.h) are not declared here. The whole of [MS-ERREF] 2.2, the NERR_* range
// included, lives in
// github.com/TheManticoreProject/Manticore/windows/errors/win32 as the WIN32_ERROR type,
// and a subset repeated here would cover a fraction of that 2703-code table while drifting
// from it. Convert a returned status with win32.WIN32_ERROR(status) and compare against
// win32.NERR_Success and the rest.

// SyntaxID returns the msgsvc abstract syntax identifier:
// 17fdd703-1827-4e34-79d4-24a55c53bb37, version 1.0.
func SyntaxID() syntax.SyntaxID {
	return syntax.SyntaxID{
		UUID:         guid.GUID{A: 0x17fdd703, B: 0x1827, C: 0x4e34, D: 0x79d4, E: 0x24a55c53bb37},
		MajorVersion: 1,
		MinorVersion: 0,
	}
}

// OpnumToName maps each on-the-wire opnum to its method name; the single source of
// truth.
var OpnumToName = map[uint16]string{
	OpnumNetrMessageNameAdd:     "NetrMessageNameAdd",
	OpnumNetrMessageNameEnum:    "NetrMessageNameEnum",
	OpnumNetrMessageNameGetInfo: "NetrMessageNameGetInfo",
	OpnumNetrMessageNameDel:     "NetrMessageNameDel",
}

// NameToOpnum is the reverse of OpnumToName, built at init so the two never drift.
var NameToOpnum = func() map[string]uint16 {
	m := make(map[string]uint16, len(OpnumToName))
	for op, name := range OpnumToName {
		m[name] = op
	}
	return m
}()
