// Package rpcinterface_82273fdce32a18c33f78827929dc23ea_0_0 is the descriptor for the eventlog RPC interface, abstract
// syntax 82273fdc-e32a-18c3-3f78-827929dc23ea version 0.0 ([MS-EVEN]).
//
// This package holds only the interface-level descriptor (abstract syntax,
// transport endpoint, opnums, and opnum<->name maps). The NDR wire types live in
// windows/protocols/ms-even and the method stubs in functions; both depend on this
// package, never the reverse.
package rpcinterface_82273fdce32a18c33f78827929dc23ea_0_0

// IDL source: [MS-EVEN] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-even/0d0bee9c-dac5-46d9-b19b-2087826c02db
// A fetched copy is kept at ms-even.idl in the interface directory.

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
	"github.com/TheManticoreProject/Manticore/windows/errors/win32"
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

// The methods are declared to return NTSTATUS, but the EventLog Remoting Protocol
// reports failures as Win32 error codes ([MS-EVEN] 3.1.4, [MS-ERREF] 2.2). Those
// Win32 codes are not declared here: the whole of [MS-ERREF] 2.2 lives in
// github.com/TheManticoreProject/Manticore/windows/errors/win32 as the WIN32_ERROR
// type, and a subset repeated here would cover a fraction of that 2703-code table
// while drifting from it. Convert a returned status with win32.WIN32_ERROR(status)
// and compare against win32.ERROR_SUCCESS and the rest.
//
// The two NTSTATUS values below are the exception. A server may fail a read with
// them directly, and they belong to [MS-ERREF] 2.3, which the Win32 table does not
// cover: it has no row for 0xC0000023 or 0xC000000D, so it renders both as hex.
// These stay declared here and StatusString decodes them before deferring to the
// shared table for everything else.
const (
	StatusBufferTooSmall   uint32 = 0xC0000023 // STATUS_BUFFER_TOO_SMALL, an [MS-ERREF] 2.3 NTSTATUS with no [MS-ERREF] 2.2 row
	StatusInvalidParameter uint32 = 0xC000000D // STATUS_INVALID_PARAMETER, an [MS-ERREF] 2.3 NTSTATUS with no [MS-ERREF] 2.2 row
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

// StatusString names the two NTSTATUS values this interface declares locally, which
// [MS-ERREF] 2.2 has no row for, and defers every other status to the shared
// [MS-ERREF] 2.2 table, which names each code the specification defines and renders
// hex only for values it does not.
func StatusString(status uint32) string {
	switch status {
	case StatusBufferTooSmall:
		return "STATUS_BUFFER_TOO_SMALL"
	case StatusInvalidParameter:
		return "STATUS_INVALID_PARAMETER"
	default:
		return win32.WIN32_ERROR(status).String()
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
