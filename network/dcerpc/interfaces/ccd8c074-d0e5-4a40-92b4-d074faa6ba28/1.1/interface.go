// Package rpcinterface_ccd8c074d0e54a4092b4d074faa6ba28_1_1 is the descriptor for the
// Witness (Service Witness Protocol) RPC interface, abstract syntax
// ccd8c074-d0e5-4a40-92b4-d074faa6ba28 version 1.1 ([MS-SWN]).
//
// An RPC interface is identified by its UUID and version, never by the transport it is
// reached over: the directory is named after the UUID with the version in the nested
// <maj>.<min>/ directory.
//
// This package holds only the interface-level descriptor (abstract syntax, transport
// endpoint, opnums, opnum<->name maps, status/version constants). NDR types live in the
// windows/protocols/ms-swn package (imported as msswn) and method stubs in functions;
// both depend on this package, never the reverse.
//
// Transport ([MS-SWN] 2.1): the Service Witness Protocol is reached over RPC dynamic
// endpoints, protocol sequence ncacn_ip_tcp (RPC over TCP/IP), with the endpoint resolved
// through the RPC endpoint mapper — it is NOT exposed over a named pipe. The RPC server
// allows any user to bind and uses [MS-RPCE] to obtain the caller identity for per-method
// access checks. PipeName is therefore empty.
package rpcinterface_ccd8c074d0e54a4092b4d074faa6ba28_1_1

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// PipeName is unused for this interface: [MS-SWN] 2.1 uses ncacn_ip_tcp with a dynamic
// endpoint (endpoint-mapper resolved), not a named pipe. It is retained, empty, for
// descriptor uniformity across interfaces.
const PipeName = ``

// Opnums for the on-the-wire methods ([MS-SWN] 3.1.4). Opnums 4 and 5 (the *Ex methods)
// are only applicable to Witness protocol version 2.
const (
	OpnumWitnessrGetInterfaceList uint16 = 0
	OpnumWitnessrRegister         uint16 = 1
	OpnumWitnessrUnRegister       uint16 = 2
	OpnumWitnessrAsyncNotify      uint16 = 3
	OpnumWitnessrRegisterEx       uint16 = 4
	OpnumWitnessrUnRegisterEx     uint16 = 5
)

// Witness protocol versions ([MS-SWN] 2.2.2.3). The Version parameter of WitnessrRegister
// / WitnessrRegisterEx MUST carry one of these values; anything else makes the server
// return ERROR_REVISION_MISMATCH.
const (
	WitnessVersionV1 uint32 = 0x00010001 // Witness protocol version 1
	WitnessVersionV2 uint32 = 0x00020000 // Witness protocol version 2 (enables the *Ex methods)
)

// Status codes returned by this interface. Every Witnessr* method returns a Win32 error
// code ([MS-ERREF] 2.2, transmitted as the DWORD return value); ERROR_SUCCESS (0)
// indicates success. These are the codes [MS-SWN] (section 3.1.4) documents its methods
// returning.
const (
	StatusSuccess           uint32 = 0x00000000 // ERROR_SUCCESS
	StatusAccessDenied      uint32 = 0x00000005 // ERROR_ACCESS_DENIED
	StatusInvalidParameter  uint32 = 0x00000057 // ERROR_INVALID_PARAMETER
	StatusNotFound          uint32 = 0x00000490 // ERROR_NOT_FOUND
	StatusRevisionMismatch  uint32 = 0x0000051A // ERROR_REVISION_MISMATCH
	StatusNoSystemResources uint32 = 0x000005AA // ERROR_NO_SYSTEM_RESOURCES
	StatusInvalidState      uint32 = 0x0000139F // ERROR_INVALID_STATE
)

// SyntaxID returns the Witness abstract syntax identifier:
// ccd8c074-d0e5-4a40-92b4-d074faa6ba28, version 1.1.
func SyntaxID() syntax.SyntaxID {
	return syntax.SyntaxID{
		UUID:         guid.GUID{A: 0xccd8c074, B: 0xd0e5, C: 0x4a40, D: 0x92b4, E: 0xd074faa6ba28},
		MajorVersion: 1,
		MinorVersion: 1,
	}
}

// StatusString returns a mnemonic for the documented status codes, otherwise the
// hex value.
func StatusString(status uint32) string {
	switch status {
	case StatusSuccess:
		return "ERROR_SUCCESS"
	case StatusAccessDenied:
		return "ERROR_ACCESS_DENIED"
	case StatusInvalidParameter:
		return "ERROR_INVALID_PARAMETER"
	case StatusNotFound:
		return "ERROR_NOT_FOUND"
	case StatusRevisionMismatch:
		return "ERROR_REVISION_MISMATCH"
	case StatusNoSystemResources:
		return "ERROR_NO_SYSTEM_RESOURCES"
	case StatusInvalidState:
		return "ERROR_INVALID_STATE"
	default:
		return fmt.Sprintf("0x%08x", status)
	}
}

// OpnumToName maps each on-the-wire opnum to its method name; the single source of
// truth.
var OpnumToName = map[uint16]string{
	OpnumWitnessrGetInterfaceList: "WitnessrGetInterfaceList",
	OpnumWitnessrRegister:         "WitnessrRegister",
	OpnumWitnessrUnRegister:       "WitnessrUnRegister",
	OpnumWitnessrAsyncNotify:      "WitnessrAsyncNotify",
	OpnumWitnessrRegisterEx:       "WitnessrRegisterEx",
	OpnumWitnessrUnRegisterEx:     "WitnessrUnRegisterEx",
}

// NameToOpnum is the reverse of OpnumToName, built at init so the two never drift.
var NameToOpnum = func() map[string]uint16 {
	m := make(map[string]uint16, len(OpnumToName))
	for op, name := range OpnumToName {
		m[name] = op
	}
	return m
}()
