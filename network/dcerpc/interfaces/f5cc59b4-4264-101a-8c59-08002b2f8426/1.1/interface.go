// Package rpcinterface_f5cc59b44264101a8c5908002b2f8426_1_1 is the descriptor for the frsrpc RPC interface, abstract
// syntax f5cc59b4-4264-101a-8c59-08002b2f8426 version 1.1 ([MS-FRS1]).
//
// This package holds only the interface-level descriptor (abstract syntax,
// transport endpoint, opnums, and opnum<->name maps). The NDR wire types live in
// windows/protocols/ms-frs1 and the method stubs in functions; both depend on this
// package, never the reverse.
package rpcinterface_f5cc59b44264101a8c5908002b2f8426_1_1

// IDL source: [MS-FRS1] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-frs1/dd60a0d9-176a-46f4-9904-000172041b92
// A fetched copy is kept at ms-frs1.idl in the interface directory.

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// PipeName is empty: FRS has no named-pipe endpoint. Both FRS interfaces use only the
// ncacn_ip_tcp protocol sequence over a dynamic endpoint assigned by the RPC endpoint
// mapper (RPCSS, port 135), optionally pinned to a static TCP port ([MS-FRS1] 2.1).
const PipeName = ``

// Opnums for the on-the-wire methods. Opnums 4, 5, 6, 7, 8, 9, 10 are "not used on the wire"
// and are omitted.
const (
	OpnumFrsRpcSendCommPkt           uint16 = 0
	OpnumFrsRpcVerifyPromotionParent uint16 = 1
	OpnumFrsRpcStartPromotionParent  uint16 = 2
	OpnumFrsNOP                      uint16 = 3
)

// The status codes this interface returns are not declared here. FRS methods return a
// Win32 error code ([MS-ERREF] 2.2): 0 on success, and all nonzero values are equivalent
// failures unless otherwise specified ([MS-FRS1] 3.3.4); FrsRpcVerifyPromotionParent
// always returns ERROR_CALL_NOT_IMPLEMENTED ([MS-FRS1] 3.3.4.2). The whole of
// [MS-ERREF] 2.2 lives in
// github.com/TheManticoreProject/Manticore/windows/errors/win32 as the WIN32_ERROR type,
// and a subset repeated here would cover a fraction of that 2703-code table while
// drifting from it. Convert a returned status with win32.WIN32_ERROR(status) and compare
// against win32.ERROR_SUCCESS and the rest.
//
// Each of the three codes this descriptor used to declare has an [MS-ERREF] 2.2 row
// under the name the specification uses, so none of them stays local: StatusSuccess is
// win32.ERROR_SUCCESS (0x00000000), ErrorAccessDenied is win32.ERROR_ACCESS_DENIED
// (0x00000005) and ErrorCallNotImplemented is win32.ERROR_CALL_NOT_IMPLEMENTED
// (0x00000078). StatusString went with them.
//
// The FRS_ERROR_* family [MS-FRS2] defines at 0x23xx is a different set of values that
// [MS-ERREF] 2.2 has no rows for at all, and it is unrelated to the FRS_ERR_* NtFrs
// service errors the shared table does carry at 0x00001F41..0x00001F51.

// SyntaxID returns the frsrpc abstract syntax identifier:
// f5cc59b4-4264-101a-8c59-08002b2f8426, version 1.1.
func SyntaxID() syntax.SyntaxID {
	return syntax.SyntaxID{
		UUID:         guid.GUID{A: 0xf5cc59b4, B: 0x4264, C: 0x101a, D: 0x8c59, E: 0x08002b2f8426},
		MajorVersion: 1,
		MinorVersion: 1,
	}
}

// OpnumToName maps each on-the-wire opnum to its method name; the single source of
// truth.
var OpnumToName = map[uint16]string{
	OpnumFrsRpcSendCommPkt:           "FrsRpcSendCommPkt",
	OpnumFrsRpcVerifyPromotionParent: "FrsRpcVerifyPromotionParent",
	OpnumFrsRpcStartPromotionParent:  "FrsRpcStartPromotionParent",
	OpnumFrsNOP:                      "FrsNOP",
}

// NameToOpnum is the reverse of OpnumToName, built at init so the two never drift.
var NameToOpnum = func() map[string]uint16 {
	m := make(map[string]uint16, len(OpnumToName))
	for op, name := range OpnumToName {
		m[name] = op
	}
	return m
}()
