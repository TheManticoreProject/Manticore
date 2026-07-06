// Package rpcinterface_0b6edbfa4a244fc68a23942b1eca65d1_1_0 is the descriptor for the IRPCAsyncNotify RPC interface, abstract
// syntax 0b6edbfa-4a24-4fc6-8a23-942b1eca65d1 version 1.0 ([MS-PAN]).
//
// IRPCAsyncNotify lets a print client register for print status notifications, open
// unidirectional/bidirectional notification channels, and exchange notification data
// with the server ([MS-PAN] 3.1). Its methods take the remote-object context handle
// produced by the IRPCRemoteObject interface (ae33069b-...).
package rpcinterface_0b6edbfa4a244fc68a23942b1eca65d1_1_0

// IDL source: [MS-PAN] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-pan/3161e1b8-098f-4f42-8a58-7e342114b643
// A fetched copy is kept at ms-pan.idl in the interface directory.

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// PipeName is empty: IRPCAsyncNotify has no named-pipe endpoint. Per [MS-PAN] section
// 2.1 the interface is bound over RPC-on-TCP (ncacn_ip_tcp) at an RPC dynamic endpoint
// assigned by the endpoint mapper ([C706] Part 4), located by the interface UUID rather
// than a well-known port or pipe.
const PipeName = ``

// Opnums for the on-the-wire methods. Opnum 2 (Opnum2NotUsedOnWire) is "not used on the
// wire" and is omitted.
const (
	OpnumIRPCAsyncNotify_RegisterClient              uint16 = 0
	OpnumIRPCAsyncNotify_UnregisterClient            uint16 = 1
	OpnumIRPCAsyncNotify_GetNewChannel               uint16 = 3
	OpnumIRPCAsyncNotify_GetNotificationSendResponse uint16 = 4
	OpnumIRPCAsyncNotify_GetNotification             uint16 = 5
	OpnumIRPCAsyncNotify_CloseChannel                uint16 = 6
)

// Status codes. IRPCAsyncNotify methods return an HRESULT: ZERO (S_OK, 0x00000000) on
// success, or a common [MS-ERREF] HRESULT / one of the protocol-specific values below on
// failure ([MS-PAN] 3.1.4). The client SHOULD treat all error return values the same,
// except where noted in the per-method processing rules.
const (
	StatusSuccess uint32 = 0x00000000 // S_OK

	// Protocol-specific error values ([MS-PAN] 3.1.4.1 IRPCAsyncNotify_RegisterClient).
	ErrorAccessDenied     uint32 = 0x80070005 // E_ACCESSDENIED: caller not authorized to register
	ErrorNotEnoughMemory  uint32 = 0x8007000E // E_OUTOFMEMORY: no memory for the new registration
	ErrorRegistrationFull uint32 = 0x80070015 // HRESULT_FROM_WIN32(ERROR_NOT_READY): registration limit reached
	ErrorInvalidName      uint32 = 0x8007007B // HRESULT_FROM_WIN32(ERROR_INVALID_NAME): pName format invalid
)

// Access rights checked by the server when authorizing a registration
// ([MS-PAN] 3.1.4.1). These combine the standard [MS-DTYP] ACCESS_MASK bits with
// printing-specific bits.
const (
	SERVER_ALL_ACCESS  uint32 = 0x000F0003 // administer/enumerate print servers
	PRINTER_ALL_ACCESS uint32 = 0x000F000C // basic/administrative use of print queues
)

// SyntaxID returns the IRPCAsyncNotify abstract syntax identifier:
// 0b6edbfa-4a24-4fc6-8a23-942b1eca65d1, version 1.0.
func SyntaxID() syntax.SyntaxID {
	return syntax.SyntaxID{
		UUID:         guid.GUID{A: 0x0b6edbfa, B: 0x4a24, C: 0x4fc6, D: 0x8a23, E: 0x942b1eca65d1},
		MajorVersion: 1,
		MinorVersion: 0,
	}
}

// StatusString returns a mnemonic for the documented status codes, otherwise the
// hex value.
func StatusString(status uint32) string {
	switch status {
	case StatusSuccess:
		return "S_OK"
	case ErrorAccessDenied:
		return "E_ACCESSDENIED"
	case ErrorNotEnoughMemory:
		return "E_OUTOFMEMORY"
	case ErrorRegistrationFull:
		return "HRESULT_FROM_WIN32(ERROR_NOT_READY)"
	case ErrorInvalidName:
		return "HRESULT_FROM_WIN32(ERROR_INVALID_NAME)"
	default:
		return fmt.Sprintf("0x%08x", status)
	}
}

// OpnumToName maps each on-the-wire opnum to its method name; the single source of
// truth.
var OpnumToName = map[uint16]string{
	OpnumIRPCAsyncNotify_RegisterClient:              "IRPCAsyncNotify_RegisterClient",
	OpnumIRPCAsyncNotify_UnregisterClient:            "IRPCAsyncNotify_UnregisterClient",
	OpnumIRPCAsyncNotify_GetNewChannel:               "IRPCAsyncNotify_GetNewChannel",
	OpnumIRPCAsyncNotify_GetNotificationSendResponse: "IRPCAsyncNotify_GetNotificationSendResponse",
	OpnumIRPCAsyncNotify_GetNotification:             "IRPCAsyncNotify_GetNotification",
	OpnumIRPCAsyncNotify_CloseChannel:                "IRPCAsyncNotify_CloseChannel",
}

// NameToOpnum is the reverse of OpnumToName, built at init so the two never drift.
var NameToOpnum = func() map[string]uint16 {
	m := make(map[string]uint16, len(OpnumToName))
	for op, name := range OpnumToName {
		m[name] = op
	}
	return m
}()
