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

// netrShareSetInfoRequest is the [in] parameter set of NetrShareSetInfo: the optional
// server name, the [in,string] (ref) share name, the info level, the inline
// [in, switch_is(Level)] SHARE_INFO union (its Tag is set to Level before marshalling),
// and the optional [in,out,unique] ParmErr pointer.
type netrShareSetInfoRequest struct {
	ServerName *ndr.WSTR `ndr:"unique"`
	NetName    ndr.WSTR
	Level      ndr.DWORD
	ShareInfo  mssrvs.SHARE_INFO
	ParmErr    *ndr.DWORD `ndr:"unique"`
}

func (*netrShareSetInfoRequest) Opnum() uint16 {
	return srvsvc.OpnumNetrShareSetInfo
}

// netrShareSetInfoResponse is the reply: the [in,out,unique] ParmErr pointer and the
// NET_API_STATUS return value.
type netrShareSetInfoResponse struct {
	ParmErr *ndr.DWORD `ndr:"unique"`
	Status  ndr.DWORD  `ndr:"retval"`
}

// NetrShareSetInfo calls NetrShareSetInfo (opnum 17), setting information about a share
// ([MS-SRVS] 3.1.4.11). The union discriminant is set to level before marshalling.
func NetrShareSetInfo(rpc ndr.Invoker, serverName string, netName string, level ndr.DWORD, shareInfo mssrvs.SHARE_INFO, parmErr *ndr.DWORD) (*ndr.DWORD, error) {
	shareInfo.Tag = level
	req := &netrShareSetInfoRequest{
		ServerName: optWStr(serverName),
		NetName:    ndr.WSTR(netName),
		Level:      level,
		ShareInfo:  shareInfo,
		ParmErr:    parmErr,
	}
	var resp netrShareSetInfoResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return nil, fmt.Errorf("NetrShareSetInfo: %w", err)
	}
	if uint32(resp.Status) != srvsvc.NERR_Success && uint32(resp.Status) != srvsvc.ERROR_MORE_DATA {
		return resp.ParmErr, fmt.Errorf("NetrShareSetInfo failed: %s", srvsvc.StatusString(uint32(resp.Status)))
	}
	return resp.ParmErr, nil
}
