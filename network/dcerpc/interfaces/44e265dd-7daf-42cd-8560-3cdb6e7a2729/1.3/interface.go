// Package rpcinterface_44e265dd7daf42cd85603cdb6e7a2729_1_3 is the descriptor for the
// Terminal Services Gateway Server Protocol (TsProxyRpcInterface) RPC interface,
// abstract syntax 44e265dd-7daf-42cd-8560-3cdb6e7a2729 version 1.3 ([MS-TSGU]).
//
// An RPC interface is identified by its UUID and version, never by the transport it is
// reached over: the directory is named after the interface UUID (with the version in the
// nested 1.3/ directory).
//
// This package holds only the interface-level descriptor: the abstract syntax
// identifier, the transport endpoint (PipeName), the opnum constants and opnum<->name
// maps, and the HRESULT/return codes with StatusString. The NDR types live in the
// protocol structures package windows/protocols/ms-tsgu (package mstsgu) and the method
// stubs in functions; both depend on this package, never the reverse.
//
// References:
//   - [MS-TSGU] Terminal Services Gateway Server Protocol:
//     https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-tsgu/ea0ac9e8-2d53-477e-ba57-b1ad01e38039
package rpcinterface_44e265dd7daf42cd85603cdb6e7a2729_1_3

// IDL source: [MS-TSGU] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-tsgu/ea0ac9e8-2d53-477e-ba57-b1ad01e38039
// A fetched copy is kept at ms-tsgu.idl in the interface directory.

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
	"github.com/TheManticoreProject/Manticore/windows/errors/hresult"
	"github.com/TheManticoreProject/Manticore/windows/errors/win32"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// PipeName is retained for descriptor uniformity, but TsProxyRpcInterface is NOT a
// named-pipe interface: per [MS-TSGU] 2.1 the protocol is carried over ncacn_http (RPC
// over HTTP, [MS-RPCH]) on TCP port 3388 for the main channel and ports 443/80 for the
// HTTP in/out channels. It is therefore empty; the transport wiring is handled by the
// RPC-over-HTTP layer, not a named pipe.
const PipeName = ``

// Opnums for the on-the-wire methods. Opnums 0, 5 are "not used on the wire"
// and are omitted.
const (
	OpnumTsProxyCreateTunnel     uint16 = 1
	OpnumTsProxyAuthorizeTunnel  uint16 = 2
	OpnumTsProxyMakeTunnelCall   uint16 = 3
	OpnumTsProxyCreateChannel    uint16 = 4
	OpnumTsProxyCloseChannel     uint16 = 6
	OpnumTsProxyCloseTunnel      uint16 = 7
	OpnumTsProxySetupReceivePipe uint16 = 8
	OpnumTsProxySendToServer     uint16 = 9
)

// Return codes ([MS-TSGU] 2.2.6 "Common Return Codes"). The values that reach a caller
// through the DWORD return of these methods come from three different namespaces, and
// only a few of them are values the shared tables can name.
//
// The E_PROXY_* HRESULTs are in FACILITY_WIN32 (7), so each wraps a Win32 code in the
// 0x59D8..0x59F9 range that Terminal Services Gateway assigns for itself. [MS-ERREF]
// 2.1.1 names five FACILITY_WIN32 values and none of these, and [MS-ERREF] 2.2 names no
// code in that range either, so neither the HRESULT table nor the HRESULT_FROM_WIN32
// derivation resolves them: hresult.HRESULT(0x800759D8).String() renders hex. They stay
// declared here and StatusString decodes them.
//
// The E_PROXY_*_CODE values are not HRESULTs at all but HRESULT_CODE, the low 16 bits on
// their own, returned over the RPC/HTTP transports. Their severity bit is clear, so
// reading one as an HRESULT would report success; they stay declared here too.
// E_PROXY_CONNECTIONABORTED_CODE is the one value of that block [MS-ERREF] 2.2 does
// name — as ERROR_CONNECTION_ABORTED, the generic socket error rather than the gateway's
// own meaning — so it keeps its [MS-TSGU] name here as well.
//
// Four values this block used to declare are gone. SEC_E_LOGON_DENIED (0x8009030C) is
// named by [MS-ERREF] 2.1.1 and resolves as hresult.SEC_E_LOGON_DENIED. ERROR_ACCESS_DENIED
// (0x00000005), ERROR_BAD_ARGUMENTS (0x000000A0) and ERROR_GRACEFUL_DISCONNECT
// (0x000004CA) are Win32 codes rather than HRESULTs — an HRESULT reading of any of them
// is a meaningless success-severity value — and [MS-ERREF] 2.2 names all three under
// exactly those names in
// github.com/TheManticoreProject/Manticore/windows/errors/win32.
const (
	// StatusSuccess is zero, which is ERROR_SUCCESS as a Win32 code and S_OK as an
	// HRESULT. The method stubs compare the returned status against it rather than
	// against hresult.HRESULT.IsSuccess, which is a range and would accept every
	// success-severity value, the HRESULT_CODE block below included.
	StatusSuccess uint32 = 0x00000000

	// HRESULTs defined by [MS-TSGU] itself, in FACILITY_WIN32 over a Win32 code range
	// [MS-ERREF] does not name.
	E_PROXY_INTERNALERROR                       uint32 = 0x800759D8
	E_PROXY_RAP_ACCESSDENIED                    uint32 = 0x800759DA
	E_PROXY_NAP_ACCESSDENIED                    uint32 = 0x800759DB
	E_PROXY_ALREADYDISCONNECTED                 uint32 = 0x800759DF
	E_PROXY_CAPABILITYMISMATCH                  uint32 = 0x800759E9
	E_PROXY_QUARANTINE_ACCESSDENIED             uint32 = 0x800759ED
	E_PROXY_NOCERTAVAILABLE                     uint32 = 0x800759EE
	E_PROXY_COOKIE_BADPACKET                    uint32 = 0x800759F7
	E_PROXY_COOKIE_AUTHENTICATION_ACCESS_DENIED uint32 = 0x800759F8
	E_PROXY_UNSUPPORTED_AUTHENTICATION_METHOD   uint32 = 0x800759F9

	// DWORD (HRESULT_CODE) values returned only over the RPC/HTTP transports, chiefly by
	// TsProxySetupReceivePipe and TsProxySendToServer ([MS-TSGU] 2.2.6).
	E_PROXY_CONNECTIONABORTED_CODE       uint32 = 0x000004D4
	E_PROXY_INTERNALERROR_CODE           uint32 = 0x000059D8
	E_PROXY_TS_CONNECTFAILED_CODE        uint32 = 0x000059DD
	E_PROXY_MAXCONNECTIONSREACHED_CODE   uint32 = 0x000059E6
	E_PROXY_NOTSUPPORTED_CODE            uint32 = 0x000059E8
	E_PROXY_SESSIONTIMEOUT_CODE          uint32 = 0x000059F6
	E_PROXY_REAUTH_AUTHN_FAILED_CODE     uint32 = 0x000059FA
	E_PROXY_REAUTH_CAP_FAILED_CODE       uint32 = 0x000059FB
	E_PROXY_REAUTH_RAP_FAILED_CODE       uint32 = 0x000059FC
	E_PROXY_SDR_NOT_SUPPORTED_BY_TS_CODE uint32 = 0x000059FD
	E_PROXY_REAUTH_NAP_FAILED_CODE       uint32 = 0x00005A00
)

// SyntaxID returns the TsProxyRpcInterface abstract syntax identifier:
// 44e265dd-7daf-42cd-8560-3cdb6e7a2729, version 1.3.
func SyntaxID() syntax.SyntaxID {
	return syntax.SyntaxID{
		UUID:         guid.GUID{A: 0x44e265dd, B: 0x7daf, C: 0x42cd, D: 0x8560, E: 0x3cdb6e7a2729},
		MajorVersion: 1,
		MinorVersion: 3,
	}
}

// StatusString names the status values [MS-TSGU] defines for itself — the E_PROXY_*
// HRESULTs, whose FACILITY_WIN32 code [MS-ERREF] leaves unnamed, and the HRESULT_CODE
// values the receive-pipe and send-to-server methods return — and defers every other
// value to the shared tables: the Win32 codes of [MS-TSGU] 2.2.6 to [MS-ERREF] 2.2, and
// anything else to the [MS-ERREF] 2.1.1 HRESULT table, which names each value the
// specification defines and renders hex for the rest.
func StatusString(status uint32) string {
	switch status {
	case StatusSuccess:
		return "ERROR_SUCCESS"
	case E_PROXY_INTERNALERROR:
		return "E_PROXY_INTERNALERROR"
	case E_PROXY_RAP_ACCESSDENIED:
		return "E_PROXY_RAP_ACCESSDENIED"
	case E_PROXY_NAP_ACCESSDENIED:
		return "E_PROXY_NAP_ACCESSDENIED"
	case E_PROXY_ALREADYDISCONNECTED:
		return "E_PROXY_ALREADYDISCONNECTED"
	case E_PROXY_CAPABILITYMISMATCH:
		return "E_PROXY_CAPABILITYMISMATCH"
	case E_PROXY_QUARANTINE_ACCESSDENIED:
		return "E_PROXY_QUARANTINE_ACCESSDENIED"
	case E_PROXY_NOCERTAVAILABLE:
		return "E_PROXY_NOCERTAVAILABLE"
	case E_PROXY_COOKIE_BADPACKET:
		return "E_PROXY_COOKIE_BADPACKET"
	case E_PROXY_COOKIE_AUTHENTICATION_ACCESS_DENIED:
		return "E_PROXY_COOKIE_AUTHENTICATION_ACCESS_DENIED"
	case E_PROXY_UNSUPPORTED_AUTHENTICATION_METHOD:
		return "E_PROXY_UNSUPPORTED_AUTHENTICATION_METHOD"
	case E_PROXY_CONNECTIONABORTED_CODE:
		return "E_PROXY_CONNECTIONABORTED"
	case E_PROXY_INTERNALERROR_CODE:
		return "E_PROXY_INTERNALERROR (HRESULT_CODE)"
	case E_PROXY_TS_CONNECTFAILED_CODE:
		return "E_PROXY_TS_CONNECTFAILED"
	case E_PROXY_MAXCONNECTIONSREACHED_CODE:
		return "E_PROXY_MAXCONNECTIONSREACHED"
	case E_PROXY_NOTSUPPORTED_CODE:
		return "E_PROXY_NOTSUPPORTED"
	case E_PROXY_SESSIONTIMEOUT_CODE:
		return "E_PROXY_SESSIONTIMEOUT"
	case E_PROXY_REAUTH_AUTHN_FAILED_CODE:
		return "E_PROXY_REAUTH_AUTHN_FAILED"
	case E_PROXY_REAUTH_CAP_FAILED_CODE:
		return "E_PROXY_REAUTH_CAP_FAILED"
	case E_PROXY_REAUTH_RAP_FAILED_CODE:
		return "E_PROXY_REAUTH_RAP_FAILED"
	case E_PROXY_SDR_NOT_SUPPORTED_BY_TS_CODE:
		return "E_PROXY_SDR_NOT_SUPPORTED_BY_TS"
	case E_PROXY_REAUTH_NAP_FAILED_CODE:
		return "E_PROXY_REAUTH_NAP_FAILED"
	case uint32(win32.ERROR_ACCESS_DENIED), uint32(win32.ERROR_BAD_ARGUMENTS), uint32(win32.ERROR_GRACEFUL_DISCONNECT):
		return win32.WIN32_ERROR(status).String()
	default:
		return hresult.HRESULT(status).String()
	}
}

// OpnumToName maps each on-the-wire opnum to its method name; the single source of
// truth.
var OpnumToName = map[uint16]string{
	OpnumTsProxyCreateTunnel:     "TsProxyCreateTunnel",
	OpnumTsProxyAuthorizeTunnel:  "TsProxyAuthorizeTunnel",
	OpnumTsProxyMakeTunnelCall:   "TsProxyMakeTunnelCall",
	OpnumTsProxyCreateChannel:    "TsProxyCreateChannel",
	OpnumTsProxyCloseChannel:     "TsProxyCloseChannel",
	OpnumTsProxyCloseTunnel:      "TsProxyCloseTunnel",
	OpnumTsProxySetupReceivePipe: "TsProxySetupReceivePipe",
	OpnumTsProxySendToServer:     "TsProxySendToServer",
}

// NameToOpnum is the reverse of OpnumToName, built at init so the two never drift.
var NameToOpnum = func() map[string]uint16 {
	m := make(map[string]uint16, len(OpnumToName))
	for op, name := range OpnumToName {
		m[name] = op
	}
	return m
}()
