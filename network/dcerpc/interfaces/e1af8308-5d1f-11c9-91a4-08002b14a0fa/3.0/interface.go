// Package rpcinterface_e1af83085d1f11c991a408002b14a0fa_3_0 is the descriptor for the
// endpoint mapper (ept, "epmapper") RPC interface, abstract syntax
// e1af8308-5d1f-11c9-91a4-08002b14a0fa version 3.0 ([C706] Appendix O, [MS-RPCE]).
//
// The endpoint mapper resolves an interface UUID and version to a concrete transport
// endpoint (typically the dynamic TCP port a service listens on). An RPC interface is
// identified by its UUID and version, never by the named pipe it is reached over, so
// the directory is named after the UUID with the version in the nested 3.0/ directory.
//
// This package holds only the interface-level descriptor (abstract syntax, transport
// endpoint, opnums, opnum<->name maps, status constants). The protocol tower (twr_t)
// and the NDR shapes live in the structures subpackage; the ept_map stub lives in
// functions. Both depend on this package, never the reverse.
package rpcinterface_e1af83085d1f11c991a408002b14a0fa_3_0

// IDL source: [C706] — this interface is translated from and verified
// against the protocol's authoritative IDL. Authoritative IDL reference:
//   https://pubs.opengroup.org/onlinepubs/9629399/apdxo.htm
// No standalone MS-* Full IDL page exists; the reference above is authoritative.

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// PipeName is the IPC$-relative named pipe for the endpoint mapper (ncacn_np). The ept
// interface is also reachable over ncacn_ip_tcp on TCP port 135 ([MS-RPCE] 2.1).
const PipeName = `\epmapper`

// Opnums for the on-the-wire ept methods ([C706] Appendix O). Only the operations
// modelled here are listed.
const (
	// OpnumEptLookup is ept_lookup (opnum 2), which enumerates endpoint-map entries.
	OpnumEptLookup uint16 = 2
	// OpnumEptMap is ept_map (opnum 3), which resolves an interface to its bound
	// endpoints via a protocol tower.
	OpnumEptMap uint16 = 3
	// OpnumEptLookupHandleFree is ept_lookup_handle_free (opnum 4), which releases a
	// lookup context handle obtained from ept_lookup. Opnums are assigned by IDL
	// declaration order in [C706] Appendix O: ept_insert(0), ept_delete(1), ept_lookup(2),
	// ept_map(3), ept_lookup_handle_free(4), ept_inq_object(5), ept_mgmt_delete(6).
	OpnumEptLookupHandleFree uint16 = 4
)

// Status codes returned in the ept_map / ept_lookup [out] error_status_t. 0 is success;
// the ept specific codes are DCE error-status values ([C706] Appendix O / Appendix E).
// ept_lookup additionally returns ept_s_not_registered (EptStatusNotRegistered) once no
// further elements match, which the paging loop treats as a normal end of enumeration.
const (
	EptStatusSuccess       uint32 = 0x00000000 // rpc_s_ok
	EptStatusCantPerform   uint32 = 0x16c9a0d8 // ept_s_cant_perform_op
	EptStatusNotRegistered uint32 = 0x16c9a0d6 // ept_s_not_registered
	EptStatusInvalidEntry  uint32 = 0x16c9a0d7 // ept_s_invalid_entry
)

// ept_lookup inquiry_type values ([C706] Appendix O / [MS-RPCE] 2.2.1.2.4): which entries
// the endpoint mapper returns. EptInquiryAllElts enumerates the whole endpoint map and
// ignores the object/Ifid/vers_option filters.
const (
	EptInquiryAllElts     uint32 = 0x00000000 // RPC_C_EP_ALL_ELTS
	EptInquiryMatchByIf   uint32 = 0x00000001 // RPC_C_EP_MATCH_BY_IF
	EptInquiryMatchByObj  uint32 = 0x00000002 // RPC_C_EP_MATCH_BY_OBJ
	EptInquiryMatchByBoth uint32 = 0x00000003 // RPC_C_EP_MATCH_BY_BOTH
)

// ept_lookup vers_option values: the interface-version constraint applied when matching by
// interface ([MS-RPCE] 2.2.1.2.4).
const (
	EptVersAll        uint32 = 0x00000001 // RPC_C_VERS_ALL
	EptVersCompatible uint32 = 0x00000002 // RPC_C_VERS_COMPATIBLE
	EptVersExact      uint32 = 0x00000003 // RPC_C_VERS_EXACT
	EptVersMajorOnly  uint32 = 0x00000004 // RPC_C_VERS_MAJOR_ONLY
	EptVersUpto       uint32 = 0x00000005 // RPC_C_VERS_UPTO
)

// SyntaxID returns the ept abstract syntax identifier:
// e1af8308-5d1f-11c9-91a4-08002b14a0fa, version 3.0.
func SyntaxID() syntax.SyntaxID {
	return syntax.SyntaxID{
		UUID:         guid.GUID{A: 0xe1af8308, B: 0x5d1f, C: 0x11c9, D: 0x91a4, E: 0x08002b14a0fa},
		MajorVersion: 3,
		MinorVersion: 0,
	}
}

// StatusString returns a mnemonic for the documented status codes, otherwise the hex
// value.
func StatusString(status uint32) string {
	switch status {
	case EptStatusSuccess:
		return "rpc_s_ok"
	case EptStatusCantPerform:
		return "ept_s_cant_perform_op"
	case EptStatusNotRegistered:
		return "ept_s_not_registered"
	case EptStatusInvalidEntry:
		return "ept_s_invalid_entry"
	default:
		return fmt.Sprintf("0x%08x", status)
	}
}

// OpnumToName maps each modelled opnum to its method name; the single source of truth.
var OpnumToName = map[uint16]string{
	OpnumEptLookupHandleFree: "ept_lookup_handle_free",
	OpnumEptLookup:           "ept_lookup",
	OpnumEptMap:              "ept_map",
}

// NameToOpnum is the reverse of OpnumToName, built at init so the two never drift.
var NameToOpnum = func() map[string]uint16 {
	m := make(map[string]uint16, len(OpnumToName))
	for op, name := range OpnumToName {
		m[name] = op
	}
	return m
}()
