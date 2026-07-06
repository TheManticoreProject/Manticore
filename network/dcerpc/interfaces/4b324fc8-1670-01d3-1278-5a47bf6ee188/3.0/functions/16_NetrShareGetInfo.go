package functions

// IDL source: [MS-SRVS] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-srvs/77aacc74-f8f9-4b46-b2d8-bfe04a7d9c44
// A fetched copy is kept at ms-srvs.idl in the interface directory.

import (
	"fmt"

	srvsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mssrvs "github.com/TheManticoreProject/Manticore/windows/protocols/ms-srvs"
)

// netrShareGetInfoRequest is the [in] parameter set of NetrShareGetInfo: the optional
// server name, the [in,string] (ref) share name, and the info level selecting the union
// arm returned.
type netrShareGetInfoRequest struct {
	ServerName *ndr.WSTR `ndr:"unique"`
	NetName    ndr.WSTR
	Level      ndr.DWORD
}

func (*netrShareGetInfoRequest) Opnum() uint16 {
	return srvsvc.OpnumNetrShareGetInfo
}

// netrShareGetInfoResponse is the reply: the [out, switch_is(Level)] SHARE_INFO union and
// the NET_API_STATUS return value.
type netrShareGetInfoResponse struct {
	InfoStruct mssrvs.SHARE_INFO
	Status     ndr.DWORD `ndr:"retval"`
}

// NetrShareGetInfo calls NetrShareGetInfo (opnum 16), retrieving information about a share
// ([MS-SRVS] 3.1.4.10).
func NetrShareGetInfo(rpc ndr.Invoker, serverName string, netName string, level ndr.DWORD) (mssrvs.SHARE_INFO, error) {
	req := &netrShareGetInfoRequest{
		ServerName: optWStr(serverName),
		NetName:    ndr.WSTR(netName),
		Level:      level,
	}
	var resp netrShareGetInfoResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return mssrvs.SHARE_INFO{}, fmt.Errorf("NetrShareGetInfo: %w", err)
	}
	if uint32(resp.Status) != srvsvc.NERR_Success && uint32(resp.Status) != srvsvc.ERROR_MORE_DATA {
		return resp.InfoStruct, fmt.Errorf("NetrShareGetInfo failed: %s", srvsvc.StatusString(uint32(resp.Status)))
	}
	return resp.InfoStruct, nil
}
