// Package rpcinterface_99fcfec45260101bbbcb00aa0021347a_0_0 is the descriptor for the
// IObjectExporter (IOXIDResolver) RPC interface, abstract syntax
// 99fcfec4-5260-101b-bbcb-00aa0021347a version 0.0 ([MS-DCOM]).
//
// IObjectExporter is the object resolver interface of the DCOM Remote Protocol: it
// resolves OXIDs to object exporter bindings (ResolveOxid/ResolveOxid2), tests server
// liveness (ServerAlive/ServerAlive2), and maintains the ping sets that keep remote
// objects alive (SimplePing/ComplexPing). Unlike the rest of MS-DCOM it is a classic,
// non-object RPC interface: it uses an explicit binding handle (handle_t) and carries no
// ORPCTHIS/ORPCTHAT, so it is modeled here while the object (ORPC) interfaces are not.
//
// An RPC interface is identified by its UUID and version, never by the transport it is
// reached over, so the directory is named after the UUID with the version in the nested
// 0.0/ directory.
package rpcinterface_99fcfec45260101bbbcb00aa0021347a_0_0

// IDL source: [MS-DCOM] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dcom/49aef5a4-f0ad-4478-abb5-cb9446dc13c6
// A fetched copy is kept at ms-dcom.idl in the interface directory.

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// PipeName is unused for this interface: IObjectExporter (the DCOM object resolver) is not
// served over a named pipe. It is reached through the RPC endpoint mapper / well-known
// endpoints ([MS-DCOM] 2.1) — over ncacn_ip_tcp on TCP port 135 and any other RPC protocol
// sequence the resolver responds to; a client discovers a working sequence with
// ServerAlive2. It is retained, empty, for descriptor uniformity across interfaces (the
// same convention as the other endpoint-mapper-resolved TCP interfaces in this tree);
// feeding an empty pipe to a named-pipe binder fails loudly rather than binding the wrong
// endpoint.
const PipeName = ``

// Opnums for the on-the-wire methods ([MS-DCOM] 3.1.2.5.1). All six opnums are used on the
// wire (IObjectExporter defines no OpnumNNotUsedOnWire placeholders).
const (
	OpnumResolveOxid  uint16 = 0
	OpnumSimplePing   uint16 = 1
	OpnumComplexPing  uint16 = 2
	OpnumServerAlive  uint16 = 3
	OpnumResolveOxid2 uint16 = 4
	OpnumServerAlive2 uint16 = 5
)

// Status codes returned by this interface. IObjectExporter methods return error_status_t,
// an RPC/Win32 status ([C706]; [MS-ERREF] 2.2). RPC_S_OK indicates success; the OR_* codes
// are the object resolver specific failures. Any other Win32/RPC status can also surface.
const (
	StatusSuccess uint32 = 0x00000000 // RPC_S_OK
	OrInvalidOxid uint32 = 0x00000776 // OR_INVALID_OXID: the object exporter identifier is not known
	OrInvalidOid  uint32 = 0x00000777 // OR_INVALID_OID: the object identifier is not known
	OrInvalidSet  uint32 = 0x00000778 // OR_INVALID_SET: the ping set identifier is not known
)

// SyntaxID returns the IObjectExporter abstract syntax identifier:
// 99fcfec4-5260-101b-bbcb-00aa0021347a, version 0.0.
func SyntaxID() syntax.SyntaxID {
	return syntax.SyntaxID{
		UUID:         guid.GUID{A: 0x99fcfec4, B: 0x5260, C: 0x101b, D: 0xbbcb, E: 0x00aa0021347a},
		MajorVersion: 0,
		MinorVersion: 0,
	}
}

// StatusString returns a mnemonic for the documented status codes, otherwise the
// hex value.
func StatusString(status uint32) string {
	switch status {
	case StatusSuccess:
		return "RPC_S_OK"
	case OrInvalidOxid:
		return "OR_INVALID_OXID"
	case OrInvalidOid:
		return "OR_INVALID_OID"
	case OrInvalidSet:
		return "OR_INVALID_SET"
	default:
		return fmt.Sprintf("0x%08x", status)
	}
}

// OpnumToName maps each on-the-wire opnum to its method name; the single source of
// truth.
var OpnumToName = map[uint16]string{
	OpnumResolveOxid:  "ResolveOxid",
	OpnumSimplePing:   "SimplePing",
	OpnumComplexPing:  "ComplexPing",
	OpnumServerAlive:  "ServerAlive",
	OpnumResolveOxid2: "ResolveOxid2",
	OpnumServerAlive2: "ServerAlive2",
}

// NameToOpnum is the reverse of OpnumToName, built at init so the two never drift.
var NameToOpnum = func() map[string]uint16 {
	m := make(map[string]uint16, len(OpnumToName))
	for op, name := range OpnumToName {
		m[name] = op
	}
	return m
}()
