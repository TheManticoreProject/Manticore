// Package rpcinterface_5a7b91f8ff0011d0a9b200c04fb6e6fc_1_0 is the descriptor for the msgsvcsend RPC interface, abstract
// syntax 5a7b91f8-ff00-11d0-a9b2-00c04fb6e6fc version 1.0 ([MS-MSRP]).
//
// The PipeName, the NET_API_STATUS code table, and doc comments are not derivable
// from the IDL and were filled in by hand from [MS-MSRP] and [MS-ERREF].
package rpcinterface_5a7b91f8ff0011d0a9b200c04fb6e6fc_1_0

// IDL source: [MS-MSRP] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-msrp/181965ff-fab4-4ad4-a8d7-16b444cc4e66
// A fetched copy is kept at ms-msrp.idl in the interface directory.

import (
	"fmt"

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

// NET_API_STATUS / error_status_t codes returned by this interface ([MS-MSRP] 3.2.4.1,
// [MS-ERREF], lmerr.h). NERR_* values are relative to NERR_BASE (2100 / 0x834).
const (
	StatusSuccess          uint32 = 0x00000000 // ERROR_SUCCESS
	ErrorAccessDenied      uint32 = 0x00000005 // ERROR_ACCESS_DENIED
	ErrorNotSupported      uint32 = 0x00000032 // ERROR_NOT_SUPPORTED
	ErrorInvalidParameter  uint32 = 0x00000057 // ERROR_INVALID_PARAMETER
	NerrNetworkError       uint32 = 0x00000858 // NERR_NetworkError
	NerrNameNotFound       uint32 = 0x000008E1 // NERR_NameNotFound
	NerrGrpMsgProcessor    uint32 = 0x000008E8 // NERR_GrpMsgProcessor
	NerrPausedRemote       uint32 = 0x000008E9 // NERR_PausedRemote
	NerrBadReceive         uint32 = 0x000008EA // NERR_BadReceive
	NerrNameInUse          uint32 = 0x000008EB // NERR_NameInUse
	NerrNotLocalName       uint32 = 0x000008ED // NERR_NotLocalName
	NerrTruncatedBroadcast uint32 = 0x000008F1 // NERR_TruncatedBroadcast
	NerrDuplicateName      uint32 = 0x000008F9 // NERR_DuplicateName
)

// SyntaxID returns the msgsvcsend abstract syntax identifier:
// 5a7b91f8-ff00-11d0-a9b2-00c04fb6e6fc, version 1.0.
func SyntaxID() syntax.SyntaxID {
	return syntax.SyntaxID{
		UUID:         guid.GUID{A: 0x5a7b91f8, B: 0xff00, C: 0x11d0, D: 0xa9b2, E: 0x00c04fb6e6fc},
		MajorVersion: 1,
		MinorVersion: 0,
	}
}

// StatusString returns a mnemonic for the documented status codes, otherwise the
// hex value.
func StatusString(status uint32) string {
	switch status {
	case StatusSuccess:
		return "ERROR_SUCCESS"
	case ErrorAccessDenied:
		return "ERROR_ACCESS_DENIED"
	case ErrorNotSupported:
		return "ERROR_NOT_SUPPORTED"
	case ErrorInvalidParameter:
		return "ERROR_INVALID_PARAMETER"
	case NerrNetworkError:
		return "NERR_NetworkError"
	case NerrNameNotFound:
		return "NERR_NameNotFound"
	case NerrGrpMsgProcessor:
		return "NERR_GrpMsgProcessor"
	case NerrPausedRemote:
		return "NERR_PausedRemote"
	case NerrBadReceive:
		return "NERR_BadReceive"
	case NerrNameInUse:
		return "NERR_NameInUse"
	case NerrNotLocalName:
		return "NERR_NotLocalName"
	case NerrTruncatedBroadcast:
		return "NERR_TruncatedBroadcast"
	case NerrDuplicateName:
		return "NERR_DuplicateName"
	default:
		return fmt.Sprintf("0x%08x", status)
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
