// Package rpcinterface_17fdd70318274e3479d424a55c53bb37_1_0 is the descriptor for the msgsvc RPC interface, abstract
// syntax 17fdd703-1827-4e34-79d4-24a55c53bb37 version 1.0 ([MS-MSRP]).
//
// The PipeName, the NET_API_STATUS code table, and doc comments are not derivable
// from the IDL and were filled in by hand from [MS-MSRP] and [MS-ERREF].
package rpcinterface_17fdd70318274e3479d424a55c53bb37_1_0

// IDL source: [MS-MSRP] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-msrp/181965ff-fab4-4ad4-a8d7-16b444cc4e66
// A fetched copy is kept at ms-msrp.idl in the interface directory.

import (
	"fmt"

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

// NET_API_STATUS codes returned by this interface ([MS-MSRP] 3.1.4, [MS-ERREF],
// lmerr.h). NERR_* values are relative to NERR_BASE (2100 / 0x834).
const (
	StatusSuccess         uint32 = 0x00000000 // NERR_Success / ERROR_SUCCESS
	ErrorAccessDenied     uint32 = 0x00000005 // ERROR_ACCESS_DENIED
	ErrorNotEnoughMemory  uint32 = 0x00000008 // ERROR_NOT_ENOUGH_MEMORY
	ErrorInvalidParameter uint32 = 0x00000057 // ERROR_INVALID_PARAMETER
	ErrorInvalidName      uint32 = 0x0000007B // ERROR_INVALID_NAME
	ErrorInvalidLevel     uint32 = 0x0000007C // ERROR_INVALID_LEVEL
	NerrBufTooSmall       uint32 = 0x0000084B // NERR_BufTooSmall
	NerrNetworkError      uint32 = 0x00000858 // NERR_NetworkError
	NerrInternalError     uint32 = 0x0000085C // NERR_InternalError
	NerrAlreadyExists     uint32 = 0x000008E4 // NERR_AlreadyExists
	NerrTooManyNames      uint32 = 0x000008E5 // NERR_TooManyNames
	NerrDelComputerName   uint32 = 0x000008E6 // NERR_DelComputerName
	NerrNameInUse         uint32 = 0x000008EB // NERR_NameInUse
	NerrNotLocalName      uint32 = 0x000008ED // NERR_NotLocalName
	NerrDuplicateName     uint32 = 0x000008F9 // NERR_DuplicateName
	NerrIncompleteDel     uint32 = 0x000008FB // NERR_IncompleteDel
)

// SyntaxID returns the msgsvc abstract syntax identifier:
// 17fdd703-1827-4e34-79d4-24a55c53bb37, version 1.0.
func SyntaxID() syntax.SyntaxID {
	return syntax.SyntaxID{
		UUID:         guid.GUID{A: 0x17fdd703, B: 0x1827, C: 0x4e34, D: 0x79d4, E: 0x24a55c53bb37},
		MajorVersion: 1,
		MinorVersion: 0,
	}
}

// StatusString returns a mnemonic for the documented status codes, otherwise the
// hex value.
func StatusString(status uint32) string {
	switch status {
	case StatusSuccess:
		return "NERR_Success"
	case ErrorAccessDenied:
		return "ERROR_ACCESS_DENIED"
	case ErrorNotEnoughMemory:
		return "ERROR_NOT_ENOUGH_MEMORY"
	case ErrorInvalidParameter:
		return "ERROR_INVALID_PARAMETER"
	case ErrorInvalidName:
		return "ERROR_INVALID_NAME"
	case ErrorInvalidLevel:
		return "ERROR_INVALID_LEVEL"
	case NerrBufTooSmall:
		return "NERR_BufTooSmall"
	case NerrNetworkError:
		return "NERR_NetworkError"
	case NerrInternalError:
		return "NERR_InternalError"
	case NerrAlreadyExists:
		return "NERR_AlreadyExists"
	case NerrTooManyNames:
		return "NERR_TooManyNames"
	case NerrDelComputerName:
		return "NERR_DelComputerName"
	case NerrNameInUse:
		return "NERR_NameInUse"
	case NerrNotLocalName:
		return "NERR_NotLocalName"
	case NerrDuplicateName:
		return "NERR_DuplicateName"
	case NerrIncompleteDel:
		return "NERR_IncompleteDel"
	default:
		return fmt.Sprintf("0x%08x", status)
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
