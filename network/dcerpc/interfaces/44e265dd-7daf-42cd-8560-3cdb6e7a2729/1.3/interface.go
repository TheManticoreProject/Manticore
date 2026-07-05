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

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
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

// Return codes ([MS-TSGU] 2.2.6 "Common Return Codes"). The HRESULT forms are returned
// by the TsProxy* methods that return HRESULT; StatusSuccess (ERROR_SUCCESS) is 0.
const (
	// StatusSuccess is ERROR_SUCCESS: the requested operation succeeded.
	StatusSuccess uint32 = 0x00000000

	// HRESULT values defined by [MS-TSGU].
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
	SEC_E_LOGON_DENIED                          uint32 = 0x8009030C

	// Win32 error codes also returned by the HRESULT-returning methods.
	ERROR_ACCESS_DENIED       uint32 = 0x00000005
	ERROR_BAD_ARGUMENTS       uint32 = 0x000000A0
	ERROR_GRACEFUL_DISCONNECT uint32 = 0x000004CA

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

// StatusString returns a mnemonic for the documented status codes, otherwise the
// hex value.
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
	case SEC_E_LOGON_DENIED:
		return "SEC_E_LOGON_DENIED"
	case ERROR_ACCESS_DENIED:
		return "ERROR_ACCESS_DENIED"
	case ERROR_BAD_ARGUMENTS:
		return "ERROR_BAD_ARGUMENTS"
	case ERROR_GRACEFUL_DISCONNECT:
		return "ERROR_GRACEFUL_DISCONNECT"
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
	default:
		return fmt.Sprintf("0x%08x", status)
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
