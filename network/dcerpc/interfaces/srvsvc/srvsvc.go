// Package srvsvc implements a minimal client for the Server Service Remote Protocol
// (srvsvc, [MS-SRVS]), built on the declarative network/dcerpc/ndr API.
//
// The calls implemented here are the ones whose NDR is fully handled by the current
// codec: NetrRemoteTOD (a pointer to a fixed all-DWORD structure) and
// NetrServerGetInfo level 101 (a union whose arm is a single pointer-to-structure).
// The enumeration calls (NetrShareEnum, NetrSessionEnum) require arrays of structures
// that themselves contain pointers, which the codec does not yet marshal with the
// correct (after-the-array) referent ordering, so they are intentionally omitted.
//
// References:
//   - [MS-SRVS] Server Service Remote Protocol:
//     https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-srvs/accf23b0-0f57-441c-9185-43041f1b0ee9
//   - [MS-SRVS] NetrRemoteTOD (Opnum 28):
//     https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-srvs/6e914f10-3280-474e-81fa-71621d5c3409
//   - [MS-SRVS] NetrServerGetInfo (Opnum 21), section 3.1.4.17
//   - [MS-SRVS] 2.2.4.105 TIME_OF_DAY_INFO; 2.2.4.41 SERVER_INFO_101
package srvsvc

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/client"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// PipeName is the named pipe (IPC$-relative) for the srvsvc interface.
const PipeName = `\srvsvc`

// Opnums for the implemented methods ([MS-SRVS] section 3.1.4).
const (
	OpnumNetrServerGetInfo uint16 = 21
	OpnumNetrRemoteTOD     uint16 = 28
)

// NERR_Success is the success return value shared by these methods.
const NERR_Success uint32 = 0x00000000

// SyntaxID returns the srvsvc abstract syntax: 4b324fc8-1670-01d3-1278-5a47bf6ee188,
// version 3.0.
func SyntaxID() syntax.SyntaxID {
	return syntax.SyntaxID{
		UUID:         guid.GUID{A: 0x4b324fc8, B: 0x1670, C: 0x01d3, D: 0x1278, E: 0x5a47bf6ee188},
		MajorVersion: 3,
		MinorVersion: 0,
	}
}

// TimeOfDayInfo is TIME_OF_DAY_INFO ([MS-SRVS] 2.2.4.105): twelve 32-bit fields, no
// pointers or arrays. tod_timezone is a signed minutes-from-GMT offset on the wire.
type TimeOfDayInfo struct {
	Elapsedt  ndr.DWORD // seconds since 1970-01-01 00:00:00 GMT
	Msecs     ndr.DWORD // milliseconds since system boot
	Hours     ndr.DWORD
	Mins      ndr.DWORD
	Secs      ndr.DWORD
	Hunds     ndr.DWORD // hundredths of a second
	Timezone  ndr.DWORD // minutes from GMT (signed)
	Tinterval ndr.DWORD // clock tick interval in 0.0001 second units
	Day       ndr.DWORD
	Month     ndr.DWORD
	Year      ndr.DWORD
	Weekday   ndr.DWORD // 0 = Sunday
}

// remoteTODRequest is the [in] parameter set of NetrRemoteTOD. ServerName is ignored
// by the server, so a NULL unique pointer is sent.
type remoteTODRequest struct {
	ServerName *ndr.WSTR `ndr:"unique"`
}

func (*remoteTODRequest) Opnum() uint16 { return OpnumNetrRemoteTOD }

// remoteTODResponse is the reply: a unique pointer to the structure plus the status.
type remoteTODResponse struct {
	Buffer    *TimeOfDayInfo `ndr:"unique"`
	ErrorCode ndr.DWORD      `ndr:"retval"`
}

// RemoteTOD calls NetrRemoteTOD (opnum 28) and returns the server's time-of-day info.
func RemoteTOD(rpc *client.Client) (*TimeOfDayInfo, error) {
	var resp remoteTODResponse
	if err := rpc.Invoke(&remoteTODRequest{}, &resp); err != nil {
		return nil, fmt.Errorf("NetrRemoteTOD: %w", err)
	}
	if uint32(resp.ErrorCode) != NERR_Success {
		return nil, fmt.Errorf("NetrRemoteTOD failed: 0x%08x", uint32(resp.ErrorCode))
	}
	if resp.Buffer == nil {
		return nil, fmt.Errorf("NetrRemoteTOD: server returned a NULL buffer")
	}
	return resp.Buffer, nil
}

// ServerInfo101 is SERVER_INFO_101 ([MS-SRVS] 2.2.4.41).
type ServerInfo101 struct {
	PlatformID   ndr.DWORD
	Name         *ndr.WSTR `ndr:"unique"`
	VersionMajor ndr.DWORD
	VersionMinor ndr.DWORD
	Type         ndr.DWORD
	Comment      *ndr.WSTR `ndr:"unique"`
}

// serverGetInfoRequest is the [in] parameter set of NetrServerGetInfo.
type serverGetInfoRequest struct {
	ServerName *ndr.WSTR `ndr:"unique"`
	Level      ndr.DWORD
}

func (*serverGetInfoRequest) Opnum() uint16 { return OpnumNetrServerGetInfo }

// serverGetInfo101Response models the reply when Level is 101. The [out, switch_is]
// SERVER_INFO union is transmitted as its discriminant followed by the selected arm;
// for level 101 the arm is a unique pointer to a SERVER_INFO_101, so the union is
// modeled inline as Discriminant + Info, keeping the pointer's referent in the
// walker's deferral path.
type serverGetInfo101Response struct {
	Discriminant ndr.DWORD
	Info         *ServerInfo101 `ndr:"unique"`
	ErrorCode    ndr.DWORD      `ndr:"retval"`
}

// ServerGetInfo101 calls NetrServerGetInfo (opnum 21) at level 101 and returns the
// server's SERVER_INFO_101 (name, version, type, comment).
func ServerGetInfo101(rpc *client.Client) (*ServerInfo101, error) {
	var resp serverGetInfo101Response
	if err := rpc.Invoke(&serverGetInfoRequest{Level: 101}, &resp); err != nil {
		return nil, fmt.Errorf("NetrServerGetInfo: %w", err)
	}
	if uint32(resp.ErrorCode) != NERR_Success {
		return nil, fmt.Errorf("NetrServerGetInfo failed: 0x%08x", uint32(resp.ErrorCode))
	}
	if resp.Info == nil {
		return nil, fmt.Errorf("NetrServerGetInfo: server returned a NULL info pointer")
	}
	return resp.Info, nil
}
