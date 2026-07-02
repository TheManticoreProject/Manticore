// Package rpcinterface_82273fdce32a18c33f78827929dc23ea_0_0 is the descriptor for the eventlog RPC interface, abstract
// syntax 82273fdc-e32a-18c3-3f78-827929dc23ea version 0.0 ([MS-EVEN]).
//
// This package holds only the interface-level descriptor (abstract syntax,
// transport endpoint, opnums, opnum<->name maps, and status constants). The NDR
// wire types live in windows/protocols/ms-even and the method stubs in functions;
// both depend on this package, never the reverse.
package rpcinterface_82273fdce32a18c33f78827929dc23ea_0_0

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// PipeName is the IPC$-relative named pipe for the eventlog interface
// ([MS-EVEN] 2.1: the server listens on the \PIPE\eventlog named pipe).
const PipeName = `\eventlog`

// Opnums for the on-the-wire methods. Opnums 19, 20, 21, 23 are "not used on the wire"
// and are omitted.
const (
	OpnumElfrClearELFW             uint16 = 0
	OpnumElfrBackupELFW            uint16 = 1
	OpnumElfrCloseEL               uint16 = 2
	OpnumElfrDeregisterEventSource uint16 = 3
	OpnumElfrNumberOfRecords       uint16 = 4
	OpnumElfrOldestRecord          uint16 = 5
	OpnumElfrChangeNotify          uint16 = 6
	OpnumElfrOpenELW               uint16 = 7
	OpnumElfrRegisterEventSourceW  uint16 = 8
	OpnumElfrOpenBELW              uint16 = 9
	OpnumElfrReadELW               uint16 = 10
	OpnumElfrReportEventW          uint16 = 11
	OpnumElfrClearELFA             uint16 = 12
	OpnumElfrBackupELFA            uint16 = 13
	OpnumElfrOpenELA               uint16 = 14
	OpnumElfrRegisterEventSourceA  uint16 = 15
	OpnumElfrOpenBELA              uint16 = 16
	OpnumElfrReadELA               uint16 = 17
	OpnumElfrReportEventA          uint16 = 18
	OpnumElfrGetLogInformation     uint16 = 22
	OpnumElfrReportEventAndSourceW uint16 = 24
	OpnumElfrReportEventExW        uint16 = 25
	OpnumElfrReportEventExA        uint16 = 26
)

// Status codes returned by this interface. The methods are declared to return
// NTSTATUS, but the EventLog Remoting Protocol reports failures as Win32 error
// codes ([MS-EVEN] 3.1.4; [MS-ERREF] 2.2/2.3). Both families are listed here.
const (
	StatusSuccess uint32 = 0x00000000 // ERROR_SUCCESS

	// Win32 error codes ([MS-ERREF] 2.2).
	ErrorFileNotFound        uint32 = 0x00000002 // ERROR_FILE_NOT_FOUND
	ErrorAccessDenied        uint32 = 0x00000005 // ERROR_ACCESS_DENIED
	ErrorInvalidHandle       uint32 = 0x00000006 // ERROR_INVALID_HANDLE
	ErrorNotEnoughMemory     uint32 = 0x00000008 // ERROR_NOT_ENOUGH_MEMORY
	ErrorHandleEOF           uint32 = 0x00000026 // ERROR_HANDLE_EOF (read past the last record)
	ErrorInvalidParameter    uint32 = 0x00000057 // ERROR_INVALID_PARAMETER
	ErrorInsufficientBuffer  uint32 = 0x0000007A // ERROR_INSUFFICIENT_BUFFER
	ErrorEventlogFileCorrupt uint32 = 0x000005DC // ERROR_EVENTLOG_FILE_CORRUPT
	ErrorEventlogCantStart   uint32 = 0x000005DD // ERROR_EVENTLOG_CANT_START
	ErrorLogFileFull         uint32 = 0x000005DE // ERROR_LOG_FILE_FULL
	ErrorEventlogFileChanged uint32 = 0x000005DF // ERROR_EVENTLOG_FILE_CHANGED

	// NTSTATUS codes ([MS-ERREF] 2.3) that the server may return directly.
	StatusBufferTooSmall   uint32 = 0xC0000023 // STATUS_BUFFER_TOO_SMALL
	StatusInvalidParameter uint32 = 0xC000000D // STATUS_INVALID_PARAMETER
)

// SyntaxID returns the eventlog abstract syntax identifier:
// 82273fdc-e32a-18c3-3f78-827929dc23ea, version 0.0.
func SyntaxID() syntax.SyntaxID {
	return syntax.SyntaxID{
		UUID:         guid.GUID{A: 0x82273fdc, B: 0xe32a, C: 0x18c3, D: 0x3f78, E: 0x827929dc23ea},
		MajorVersion: 0,
		MinorVersion: 0,
	}
}

// StatusString returns a mnemonic for the documented status codes, otherwise the
// hex value.
func StatusString(status uint32) string {
	switch status {
	case StatusSuccess:
		return "ERROR_SUCCESS"
	case ErrorFileNotFound:
		return "ERROR_FILE_NOT_FOUND"
	case ErrorAccessDenied:
		return "ERROR_ACCESS_DENIED"
	case ErrorInvalidHandle:
		return "ERROR_INVALID_HANDLE"
	case ErrorNotEnoughMemory:
		return "ERROR_NOT_ENOUGH_MEMORY"
	case ErrorHandleEOF:
		return "ERROR_HANDLE_EOF"
	case ErrorInvalidParameter:
		return "ERROR_INVALID_PARAMETER"
	case ErrorInsufficientBuffer:
		return "ERROR_INSUFFICIENT_BUFFER"
	case ErrorEventlogFileCorrupt:
		return "ERROR_EVENTLOG_FILE_CORRUPT"
	case ErrorEventlogCantStart:
		return "ERROR_EVENTLOG_CANT_START"
	case ErrorLogFileFull:
		return "ERROR_LOG_FILE_FULL"
	case ErrorEventlogFileChanged:
		return "ERROR_EVENTLOG_FILE_CHANGED"
	case StatusBufferTooSmall:
		return "STATUS_BUFFER_TOO_SMALL"
	case StatusInvalidParameter:
		return "STATUS_INVALID_PARAMETER"
	default:
		return fmt.Sprintf("0x%08x", status)
	}
}

// OpnumToName maps each on-the-wire opnum to its method name; the single source of
// truth.
var OpnumToName = map[uint16]string{
	OpnumElfrClearELFW:             "ElfrClearELFW",
	OpnumElfrBackupELFW:            "ElfrBackupELFW",
	OpnumElfrCloseEL:               "ElfrCloseEL",
	OpnumElfrDeregisterEventSource: "ElfrDeregisterEventSource",
	OpnumElfrNumberOfRecords:       "ElfrNumberOfRecords",
	OpnumElfrOldestRecord:          "ElfrOldestRecord",
	OpnumElfrChangeNotify:          "ElfrChangeNotify",
	OpnumElfrOpenELW:               "ElfrOpenELW",
	OpnumElfrRegisterEventSourceW:  "ElfrRegisterEventSourceW",
	OpnumElfrOpenBELW:              "ElfrOpenBELW",
	OpnumElfrReadELW:               "ElfrReadELW",
	OpnumElfrReportEventW:          "ElfrReportEventW",
	OpnumElfrClearELFA:             "ElfrClearELFA",
	OpnumElfrBackupELFA:            "ElfrBackupELFA",
	OpnumElfrOpenELA:               "ElfrOpenELA",
	OpnumElfrRegisterEventSourceA:  "ElfrRegisterEventSourceA",
	OpnumElfrOpenBELA:              "ElfrOpenBELA",
	OpnumElfrReadELA:               "ElfrReadELA",
	OpnumElfrReportEventA:          "ElfrReportEventA",
	OpnumElfrGetLogInformation:     "ElfrGetLogInformation",
	OpnumElfrReportEventAndSourceW: "ElfrReportEventAndSourceW",
	OpnumElfrReportEventExW:        "ElfrReportEventExW",
	OpnumElfrReportEventExA:        "ElfrReportEventExA",
}

// NameToOpnum is the reverse of OpnumToName, built at init so the two never drift.
var NameToOpnum = func() map[string]uint16 {
	m := make(map[string]uint16, len(OpnumToName))
	for op, name := range OpnumToName {
		m[name] = op
	}
	return m
}()
