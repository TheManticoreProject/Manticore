// Package rpcinterface_8fb6d884238811d08c3500c04fda2795_4_1 is the descriptor for the W32Time RPC interface, abstract
// syntax 8fb6d884-2388-11d0-8c35-00c04fda2795 version 4.1 ([MS-W32T]).
//
// This package holds only the interface-level descriptor (abstract syntax, transport
// endpoints, opnums, and opnum<->name maps). The NDR types live in
// windows/protocols/ms-w32t and the method stubs in the functions subpackage; both
// depend on this package, never the reverse.
package rpcinterface_8fb6d884238811d08c3500c04fda2795_4_1

// IDL source: [MS-W32T] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-w32t/5793e908-22ef-4cea-962f-fca8a72c485a
// A fetched copy is kept at ms-w32t.idl in the interface directory.

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// PipeName is the IPC$-relative named pipe for the W32Time interface ([MS-W32T]
// section 2.1). \W32TIME serves the unauthenticated RPC interface; the authenticated
// interface is served over \W32TIME_ALT (RPC negotiates authentication on the client's
// behalf, [MS-RPCE] sections 2.2.2.11 and 5.1.1). Both reach this same abstract syntax.
const (
	PipeName    = `\W32TIME`
	PipeNameAlt = `\W32TIME_ALT`
)

// Opnums for the on-the-wire methods.
const (
	OpnumW32TimeSync                       uint16 = 0
	OpnumW32TimeGetNetlogonServiceBits     uint16 = 1
	OpnumW32TimeQueryProviderStatus        uint16 = 2
	OpnumW32TimeQuerySource                uint16 = 3
	OpnumW32TimeQueryProviderConfiguration uint16 = 4
	OpnumW32TimeQueryConfiguration         uint16 = 5
	OpnumW32TimeQueryStatus                uint16 = 6
	OpnumW32TimeLog                        uint16 = 7
)

// Every W32Time method returns a Win32 error code as its RPC return value ([MS-W32T]
// section 3.2.4): 0 (ERROR_SUCCESS) on success, otherwise one of the standard Win32
// codes of [MS-ERREF] section 2.2. Those codes are not declared here. The whole of
// [MS-ERREF] 2.2 lives in
// github.com/TheManticoreProject/Manticore/windows/errors/win32 as the WIN32_ERROR
// type, and a subset repeated here would cover a fraction of that 2703-code table
// while drifting from it. Convert a returned status with win32.WIN32_ERROR(status) and
// compare against win32.ERROR_SUCCESS and the rest. Every value this interface used to
// declare has a row in [MS-ERREF] 2.2 under the name the declaration carried, so
// nothing is kept local.

// SyntaxID returns the W32Time abstract syntax identifier:
// 8fb6d884-2388-11d0-8c35-00c04fda2795, version 4.1.
func SyntaxID() syntax.SyntaxID {
	return syntax.SyntaxID{
		UUID:         guid.GUID{A: 0x8fb6d884, B: 0x2388, C: 0x11d0, D: 0x8c35, E: 0x00c04fda2795},
		MajorVersion: 4,
		MinorVersion: 1,
	}
}

// OpnumToName maps each on-the-wire opnum to its method name; the single source of
// truth.
var OpnumToName = map[uint16]string{
	OpnumW32TimeSync:                       "W32TimeSync",
	OpnumW32TimeGetNetlogonServiceBits:     "W32TimeGetNetlogonServiceBits",
	OpnumW32TimeQueryProviderStatus:        "W32TimeQueryProviderStatus",
	OpnumW32TimeQuerySource:                "W32TimeQuerySource",
	OpnumW32TimeQueryProviderConfiguration: "W32TimeQueryProviderConfiguration",
	OpnumW32TimeQueryConfiguration:         "W32TimeQueryConfiguration",
	OpnumW32TimeQueryStatus:                "W32TimeQueryStatus",
	OpnumW32TimeLog:                        "W32TimeLog",
}

// NameToOpnum is the reverse of OpnumToName, built at init so the two never drift.
var NameToOpnum = func() map[string]uint16 {
	m := make(map[string]uint16, len(OpnumToName))
	for op, name := range OpnumToName {
		m[name] = op
	}
	return m
}()
