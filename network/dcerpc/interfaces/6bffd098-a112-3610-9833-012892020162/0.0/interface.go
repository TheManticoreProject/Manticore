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

import (
	"fmt"

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

// NET_API_STATUS codes returned by I_BrowserrQueryOtherDomains ([MS-BRWSA] 3.1.4.1;
// values from [MS-ERREF] Win32 error codes).
const (
	NERR_Success            uint32 = 0x00000000 // The operation completed successfully.
	ERROR_ACCESS_DENIED     uint32 = 0x00000005 // Access is denied.
	ERROR_NOT_ENOUGH_MEMORY uint32 = 0x00000008 // The server could not allocate enough memory.
	ERROR_INVALID_PARAMETER uint32 = 0x00000057 // A parameter is incorrect (e.g. InfoStruct or Level100 is NULL).
	ERROR_INVALID_LEVEL     uint32 = 0x0000007C // The Level member is not 100.
	ERROR_MORE_DATA         uint32 = 0x000000EA // Not all available entries were returned.
)

// SyntaxID returns the browser abstract syntax identifier:
// 6bffd098-a112-3610-9833-012892020162, version 0.0.
func SyntaxID() syntax.SyntaxID {
	return syntax.SyntaxID{
		UUID:         guid.GUID{A: 0x6bffd098, B: 0xa112, C: 0x3610, D: 0x9833, E: 0x012892020162},
		MajorVersion: 0,
		MinorVersion: 0,
	}
}

// StatusString returns a mnemonic for the documented status codes, otherwise the
// hex value.
func StatusString(status uint32) string {
	switch status {
	case NERR_Success:
		return "NERR_Success"
	case ERROR_ACCESS_DENIED:
		return "ERROR_ACCESS_DENIED"
	case ERROR_NOT_ENOUGH_MEMORY:
		return "ERROR_NOT_ENOUGH_MEMORY"
	case ERROR_INVALID_PARAMETER:
		return "ERROR_INVALID_PARAMETER"
	case ERROR_INVALID_LEVEL:
		return "ERROR_INVALID_LEVEL"
	case ERROR_MORE_DATA:
		return "ERROR_MORE_DATA"
	default:
		return fmt.Sprintf("0x%08x", status)
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
