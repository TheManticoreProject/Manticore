// Package rpcinterface_a8e0653c27444389a61d7373df8b2292_1_0 is the descriptor for the FileServerVssAgent RPC interface, abstract
// syntax a8e0653c-2744-4389-a61d-7373df8b2292 version 1.0 ([MS-FSRVP]).
//
// This package holds only the interface-level descriptor (abstract syntax, transport
// endpoint, opnums, opnum<->name maps, status codes). The PipeName ("FssagentRpc",
// [MS-FSRVP] 2.1) and the status-code table ([MS-FSRVP] 2.2.4) are not carried by the
// IDL and were filled in from the specification. NDR wire types live in
// windows/protocols/ms-fsrvp and method stubs in functions; both depend on this
// package, never the reverse.
package rpcinterface_a8e0653c27444389a61d7373df8b2292_1_0

// IDL source: [MS-FSRVP] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-fsrvp/23382633-78f1-419e-bad0-699dff0c6ef1
// A fetched copy is kept at ms-fsrvp.idl in the interface directory.

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
	"github.com/TheManticoreProject/Manticore/windows/errors/hresult"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// PipeName is the IPC$-relative named pipe for the FileServerVssAgent interface
// ([MS-FSRVP] section 2.1 Transport: the "FssagentRpc" named pipe over ncacn_np).
const PipeName = `\FssagentRpc`

// Opnums for the on-the-wire methods.
const (
	OpnumGetSupportedVersion           uint16 = 0
	OpnumSetContext                    uint16 = 1
	OpnumStartShadowCopySet            uint16 = 2
	OpnumAddToShadowCopySet            uint16 = 3
	OpnumCommitShadowCopySet           uint16 = 4
	OpnumExposeShadowCopySet           uint16 = 5
	OpnumRecoveryCompleteShadowCopySet uint16 = 6
	OpnumAbortShadowCopySet            uint16 = 7
	OpnumIsPathSupported               uint16 = 8
	OpnumIsPathShadowCopied            uint16 = 9
	OpnumGetShareMapping               uint16 = 10
	OpnumDeleteShareMapping            uint16 = 11
	OpnumPrepareShadowCopySet          uint16 = 12
)

// Status codes ([MS-FSRVP] section 2.2.4 Error Codes). FSRVP methods return an HRESULT:
// ZERO (0x00000000) on success, a few common [MS-ERREF] HRESULTs, or one of the codes
// the protocol defines for itself.
//
// The three common ones are gone from this block. E_INVALIDARG (0x80070057),
// E_ACCESSDENIED (0x80070005) and E_OUTOFMEMORY (0x8007000E) each have a row in
// [MS-ERREF] 2.1.1 under exactly that name, so they resolve through
// github.com/TheManticoreProject/Manticore/windows/errors/hresult and StatusString
// defers to it for them.
//
// The FSRVP_E_* codes below cannot follow. Six of them sit at 0x8004 23xx, which is
// FACILITY_ITF (4), the facility COM reserves for interface-specific meanings: an
// interface may define whatever it likes there, [MS-FSRVP] does exactly that, and
// [MS-ERREF] 2.1.1 has no row for any of 0x80042301, 0x80042308, 0x8004230C, 0x8004230D,
// 0x80042316 or 0x8004231B — nor for 0x80042501. FSRVP_E_WAIT_TIMEOUT and
// FSRVP_E_WAIT_FAILED are not HRESULTs at all but the Win32 wait results carried
// verbatim, and 0x00000102 has its severity bit clear, so an HRESULT reading of it
// reports success. All nine therefore stay declared here and StatusString decodes them
// before deferring to the shared table for everything else.
const (
	// StatusSuccess is zero, the value [MS-FSRVP] 2.2.4 calls ZERO and the shared table
	// calls S_OK. The method stubs compare the returned status against it rather than
	// against hresult.HRESULT.IsSuccess, which is a range: IsSuccess accepts every
	// success-severity value, FSRVP_E_WAIT_TIMEOUT among them.
	StatusSuccess uint32 = 0x00000000

	// Error codes [MS-FSRVP] defines for itself, which [MS-ERREF] 2.1.1 does not name.
	FsrvpEBadState                uint32 = 0x80042301 // FSRVP_E_BAD_STATE
	FsrvpENotSupported            uint32 = 0x8004230C // FSRVP_E_NOT_SUPPORTED
	FsrvpEObjectAlreadyExists     uint32 = 0x8004230D // FSRVP_E_OBJECT_ALREADY_EXISTS
	FsrvpEObjectNotFound          uint32 = 0x80042308 // FSRVP_E_OBJECT_NOT_FOUND
	FsrvpEShadowCopySetInProgress uint32 = 0x80042316 // FSRVP_E_SHADOW_COPY_SET_IN_PROGRESS
	FsrvpEUnsupportedContext      uint32 = 0x8004231B // FSRVP_E_UNSUPPORTED_CONTEXT
	FsrvpEShadowcopysetIdMismatch uint32 = 0x80042501 // FSRVP_E_SHADOWCOPYSET_ID_MISMATCH
	FsrvpEWaitTimeout             uint32 = 0x00000102 // FSRVP_E_WAIT_TIMEOUT, success severity
	FsrvpEWaitFailed              uint32 = 0xFFFFFFFF // FSRVP_E_WAIT_FAILED
)

// SyntaxID returns the FileServerVssAgent abstract syntax identifier:
// a8e0653c-2744-4389-a61d-7373df8b2292, version 1.0.
func SyntaxID() syntax.SyntaxID {
	return syntax.SyntaxID{
		UUID:         guid.GUID{A: 0xa8e0653c, B: 0x2744, C: 0x4389, D: 0xa61d, E: 0x7373df8b2292},
		MajorVersion: 1,
		MinorVersion: 0,
	}
}

// StatusString names the codes [MS-FSRVP] 2.2.4 defines for itself, which [MS-ERREF]
// 2.1.1 does not name — the FACILITY_ITF block the protocol fills with its own meanings,
// and the two Win32 wait results it carries verbatim — and defers every other status to
// the shared [MS-ERREF] 2.1.1 table, which names each HRESULT the specification defines
// and renders hex only for a value it does not.
func StatusString(status uint32) string {
	switch status {
	case StatusSuccess:
		return "ZERO"
	case FsrvpEBadState:
		return "FSRVP_E_BAD_STATE"
	case FsrvpENotSupported:
		return "FSRVP_E_NOT_SUPPORTED"
	case FsrvpEObjectAlreadyExists:
		return "FSRVP_E_OBJECT_ALREADY_EXISTS"
	case FsrvpEObjectNotFound:
		return "FSRVP_E_OBJECT_NOT_FOUND"
	case FsrvpEShadowCopySetInProgress:
		return "FSRVP_E_SHADOW_COPY_SET_IN_PROGRESS"
	case FsrvpEUnsupportedContext:
		return "FSRVP_E_UNSUPPORTED_CONTEXT"
	case FsrvpEShadowcopysetIdMismatch:
		return "FSRVP_E_SHADOWCOPYSET_ID_MISMATCH"
	case FsrvpEWaitTimeout:
		return "FSRVP_E_WAIT_TIMEOUT"
	case FsrvpEWaitFailed:
		return "FSRVP_E_WAIT_FAILED"
	default:
		return hresult.HRESULT(status).String()
	}
}

// OpnumToName maps each on-the-wire opnum to its method name; the single source of
// truth.
var OpnumToName = map[uint16]string{
	OpnumGetSupportedVersion:           "GetSupportedVersion",
	OpnumSetContext:                    "SetContext",
	OpnumStartShadowCopySet:            "StartShadowCopySet",
	OpnumAddToShadowCopySet:            "AddToShadowCopySet",
	OpnumCommitShadowCopySet:           "CommitShadowCopySet",
	OpnumExposeShadowCopySet:           "ExposeShadowCopySet",
	OpnumRecoveryCompleteShadowCopySet: "RecoveryCompleteShadowCopySet",
	OpnumAbortShadowCopySet:            "AbortShadowCopySet",
	OpnumIsPathSupported:               "IsPathSupported",
	OpnumIsPathShadowCopied:            "IsPathShadowCopied",
	OpnumGetShareMapping:               "GetShareMapping",
	OpnumDeleteShareMapping:            "DeleteShareMapping",
	OpnumPrepareShadowCopySet:          "PrepareShadowCopySet",
}

// NameToOpnum is the reverse of OpnumToName, built at init so the two never drift.
var NameToOpnum = func() map[string]uint16 {
	m := make(map[string]uint16, len(OpnumToName))
	for op, name := range OpnumToName {
		m[name] = op
	}
	return m
}()
