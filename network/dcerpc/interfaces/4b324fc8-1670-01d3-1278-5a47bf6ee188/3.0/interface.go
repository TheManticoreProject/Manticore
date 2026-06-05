// Package rpcinterface_4b324fc8167001d312785a47bf6ee188_3_0 is the descriptor for the
// Server Service Remote Protocol (srvsvc) RPC interface, abstract syntax
// 4b324fc8-1670-01d3-1278-5a47bf6ee188 version 3.0 ([MS-SRVS]).
//
// An RPC interface is identified by its UUID and version, never by the named pipe it is
// reached over: the "\srvsvc" pipe carries this interface, but the directory is named
// after the interface UUID (with the version in the nested 3.0/ directory).
//
// This package holds only the interface-level descriptor: the abstract syntax
// identifier, the transport endpoint (PipeName), the opnum constants and opnum<->name
// maps, and the NET_API_STATUS return codes. The NDR types live in the structures
// subpackage and the method stubs in functions; both depend on this package, never the
// reverse.
//
// References:
//   - [MS-SRVS] Server Service Remote Protocol:
//     https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-srvs/accf23b0-0f57-441c-9185-43041f1b0ee9
package rpcinterface_4b324fc8167001d312785a47bf6ee188_3_0

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// PipeName is the IPC$-relative named pipe for the srvsvc interface.
const PipeName = `\srvsvc`

// Opnums for the on-the-wire methods ([MS-SRVS] section 3.1.4). Opnums 0–7, 29, 42 and
// 47 are "not used on the wire" and are omitted.
const (
	OpnumNetrConnectionEnum           uint16 = 8
	OpnumNetrFileEnum                 uint16 = 9
	OpnumNetrFileGetInfo              uint16 = 10
	OpnumNetrFileClose                uint16 = 11
	OpnumNetrSessionEnum              uint16 = 12
	OpnumNetrSessionDel               uint16 = 13
	OpnumNetrShareAdd                 uint16 = 14
	OpnumNetrShareEnum                uint16 = 15
	OpnumNetrShareGetInfo             uint16 = 16
	OpnumNetrShareSetInfo             uint16 = 17
	OpnumNetrShareDel                 uint16 = 18
	OpnumNetrShareDelSticky           uint16 = 19
	OpnumNetrShareCheck               uint16 = 20
	OpnumNetrServerGetInfo            uint16 = 21
	OpnumNetrServerSetInfo            uint16 = 22
	OpnumNetrServerDiskEnum           uint16 = 23
	OpnumNetrServerStatisticsGet      uint16 = 24
	OpnumNetrServerTransportAdd       uint16 = 25
	OpnumNetrServerTransportEnum      uint16 = 26
	OpnumNetrServerTransportDel       uint16 = 27
	OpnumNetrRemoteTOD                uint16 = 28
	OpnumNetprPathType                uint16 = 30
	OpnumNetprPathCanonicalize        uint16 = 31
	OpnumNetprPathCompare             uint16 = 32
	OpnumNetprNameValidate            uint16 = 33
	OpnumNetprNameCanonicalize        uint16 = 34
	OpnumNetprNameCompare             uint16 = 35
	OpnumNetrShareEnumSticky          uint16 = 36
	OpnumNetrShareDelStart            uint16 = 37
	OpnumNetrShareDelCommit           uint16 = 38
	OpnumNetrpGetFileSecurity         uint16 = 39
	OpnumNetrpSetFileSecurity         uint16 = 40
	OpnumNetrServerTransportAddEx     uint16 = 41
	OpnumNetrDfsGetVersion            uint16 = 43
	OpnumNetrDfsCreateLocalPartition  uint16 = 44
	OpnumNetrDfsDeleteLocalPartition  uint16 = 45
	OpnumNetrDfsSetLocalVolumeState   uint16 = 46
	OpnumNetrDfsCreateExitPoint       uint16 = 48
	OpnumNetrDfsDeleteExitPoint       uint16 = 49
	OpnumNetrDfsModifyPrefix          uint16 = 50
	OpnumNetrDfsFixLocalVolume        uint16 = 51
	OpnumNetrDfsManagerReportSiteInfo uint16 = 52
	OpnumNetrServerTransportDelEx     uint16 = 53
	OpnumNetrServerAliasAdd           uint16 = 54
	OpnumNetrServerAliasEnum          uint16 = 55
	OpnumNetrServerAliasDel           uint16 = 56
	OpnumNetrShareDelEx               uint16 = 57
)

// NET_API_STATUS return codes ([MS-ERREF] 2.2 Win32 error codes / lmerr.h). srvsvc
// methods return a NET_API_STATUS (a 32-bit Win32 error) rather than an NTSTATUS.
const (
	NERR_Success            uint32 = 0
	ERROR_FILE_NOT_FOUND    uint32 = 2
	ERROR_ACCESS_DENIED     uint32 = 5
	ERROR_NOT_SUPPORTED     uint32 = 50
	ERROR_INVALID_PARAMETER uint32 = 87
	ERROR_INVALID_NAME      uint32 = 123
	ERROR_INVALID_LEVEL     uint32 = 124
	ERROR_MORE_DATA         uint32 = 234
	NERR_BASE               uint32 = 2100
	NERR_BufTooSmall        uint32 = 2123
	NERR_NetNameNotFound    uint32 = 2310
	NERR_DeviceNotShared    uint32 = 2311
	NERR_ClientNameNotFound uint32 = 2312
	NERR_UserNotFound       uint32 = 2221
	NERR_DuplicateShare     uint32 = 2118
)

// SyntaxID returns the srvsvc abstract syntax identifier:
// 4b324fc8-1670-01d3-1278-5a47bf6ee188, version 3.0.
func SyntaxID() syntax.SyntaxID {
	return syntax.SyntaxID{
		UUID:         guid.GUID{A: 0x4b324fc8, B: 0x1670, C: 0x01d3, D: 0x1278, E: 0x5a47bf6ee188},
		MajorVersion: 3,
		MinorVersion: 0,
	}
}

// StatusString returns a mnemonic for the documented NET_API_STATUS codes, otherwise the
// decimal value (Win32 error codes are conventionally decimal).
func StatusString(status uint32) string {
	switch status {
	case NERR_Success:
		return "NERR_Success"
	case ERROR_FILE_NOT_FOUND:
		return "ERROR_FILE_NOT_FOUND"
	case ERROR_ACCESS_DENIED:
		return "ERROR_ACCESS_DENIED"
	case ERROR_NOT_SUPPORTED:
		return "ERROR_NOT_SUPPORTED"
	case ERROR_INVALID_PARAMETER:
		return "ERROR_INVALID_PARAMETER"
	case ERROR_INVALID_NAME:
		return "ERROR_INVALID_NAME"
	case ERROR_INVALID_LEVEL:
		return "ERROR_INVALID_LEVEL"
	case ERROR_MORE_DATA:
		return "ERROR_MORE_DATA"
	case NERR_BufTooSmall:
		return "NERR_BufTooSmall"
	case NERR_NetNameNotFound:
		return "NERR_NetNameNotFound"
	case NERR_DeviceNotShared:
		return "NERR_DeviceNotShared"
	case NERR_ClientNameNotFound:
		return "NERR_ClientNameNotFound"
	case NERR_UserNotFound:
		return "NERR_UserNotFound"
	case NERR_DuplicateShare:
		return "NERR_DuplicateShare"
	default:
		return fmt.Sprintf("%d", status)
	}
}

// OpnumToName maps each on-the-wire opnum to its [MS-SRVS] method name; the single
// source of truth.
var OpnumToName = map[uint16]string{
	OpnumNetrConnectionEnum:           "NetrConnectionEnum",
	OpnumNetrFileEnum:                 "NetrFileEnum",
	OpnumNetrFileGetInfo:              "NetrFileGetInfo",
	OpnumNetrFileClose:                "NetrFileClose",
	OpnumNetrSessionEnum:              "NetrSessionEnum",
	OpnumNetrSessionDel:               "NetrSessionDel",
	OpnumNetrShareAdd:                 "NetrShareAdd",
	OpnumNetrShareEnum:                "NetrShareEnum",
	OpnumNetrShareGetInfo:             "NetrShareGetInfo",
	OpnumNetrShareSetInfo:             "NetrShareSetInfo",
	OpnumNetrShareDel:                 "NetrShareDel",
	OpnumNetrShareDelSticky:           "NetrShareDelSticky",
	OpnumNetrShareCheck:               "NetrShareCheck",
	OpnumNetrServerGetInfo:            "NetrServerGetInfo",
	OpnumNetrServerSetInfo:            "NetrServerSetInfo",
	OpnumNetrServerDiskEnum:           "NetrServerDiskEnum",
	OpnumNetrServerStatisticsGet:      "NetrServerStatisticsGet",
	OpnumNetrServerTransportAdd:       "NetrServerTransportAdd",
	OpnumNetrServerTransportEnum:      "NetrServerTransportEnum",
	OpnumNetrServerTransportDel:       "NetrServerTransportDel",
	OpnumNetrRemoteTOD:                "NetrRemoteTOD",
	OpnumNetprPathType:                "NetprPathType",
	OpnumNetprPathCanonicalize:        "NetprPathCanonicalize",
	OpnumNetprPathCompare:             "NetprPathCompare",
	OpnumNetprNameValidate:            "NetprNameValidate",
	OpnumNetprNameCanonicalize:        "NetprNameCanonicalize",
	OpnumNetprNameCompare:             "NetprNameCompare",
	OpnumNetrShareEnumSticky:          "NetrShareEnumSticky",
	OpnumNetrShareDelStart:            "NetrShareDelStart",
	OpnumNetrShareDelCommit:           "NetrShareDelCommit",
	OpnumNetrpGetFileSecurity:         "NetrpGetFileSecurity",
	OpnumNetrpSetFileSecurity:         "NetrpSetFileSecurity",
	OpnumNetrServerTransportAddEx:     "NetrServerTransportAddEx",
	OpnumNetrDfsGetVersion:            "NetrDfsGetVersion",
	OpnumNetrDfsCreateLocalPartition:  "NetrDfsCreateLocalPartition",
	OpnumNetrDfsDeleteLocalPartition:  "NetrDfsDeleteLocalPartition",
	OpnumNetrDfsSetLocalVolumeState:   "NetrDfsSetLocalVolumeState",
	OpnumNetrDfsCreateExitPoint:       "NetrDfsCreateExitPoint",
	OpnumNetrDfsDeleteExitPoint:       "NetrDfsDeleteExitPoint",
	OpnumNetrDfsModifyPrefix:          "NetrDfsModifyPrefix",
	OpnumNetrDfsFixLocalVolume:        "NetrDfsFixLocalVolume",
	OpnumNetrDfsManagerReportSiteInfo: "NetrDfsManagerReportSiteInfo",
	OpnumNetrServerTransportDelEx:     "NetrServerTransportDelEx",
	OpnumNetrServerAliasAdd:           "NetrServerAliasAdd",
	OpnumNetrServerAliasEnum:          "NetrServerAliasEnum",
	OpnumNetrServerAliasDel:           "NetrServerAliasDel",
	OpnumNetrShareDelEx:               "NetrShareDelEx",
}

// NameToOpnum is the reverse of OpnumToName, built at init so the two never drift.
var NameToOpnum = func() map[string]uint16 {
	m := make(map[string]uint16, len(OpnumToName))
	for op, name := range OpnumToName {
		m[name] = op
	}
	return m
}()
