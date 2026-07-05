package functions

import (
	"fmt"

	srvsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mssrvs "github.com/TheManticoreProject/Manticore/windows/protocols/ms-srvs"
)

// netrShareDelStartRequest is the [in] parameter set of NetrShareDelStart: the optional
// server name, the [in,string] (ref) share name, and a reserved DWORD (MUST be zero).
type netrShareDelStartRequest struct {
	ServerName *ndr.WSTR `ndr:"unique"`
	NetName    ndr.WSTR
	Reserved   ndr.DWORD
}

func (*netrShareDelStartRequest) Opnum() uint16 {
	return srvsvc.OpnumNetrShareDelStart
}

// netrShareDelStartResponse is the reply: the [out] SHARE_DEL_HANDLE context handle and the
// NET_API_STATUS return value.
type netrShareDelStartResponse struct {
	ContextHandle mssrvs.SHARE_DEL_HANDLE
	Status        ndr.DWORD `ndr:"retval"`
}

// NetrShareDelStart calls NetrShareDelStart (opnum 37), beginning a two-phase share delete
// and returning a context handle consumed by NetrShareDelCommit ([MS-SRVS] 3.1.4.16).
func NetrShareDelStart(rpc ndr.Invoker, serverName string, netName string, reserved ndr.DWORD) (mssrvs.SHARE_DEL_HANDLE, error) {
	req := &netrShareDelStartRequest{
		ServerName: optWStr(serverName),
		NetName:    ndr.WSTR(netName),
		Reserved:   reserved,
	}
	var resp netrShareDelStartResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return mssrvs.SHARE_DEL_HANDLE{}, fmt.Errorf("NetrShareDelStart: %w", err)
	}
	if uint32(resp.Status) != srvsvc.NERR_Success && uint32(resp.Status) != srvsvc.ERROR_MORE_DATA {
		return resp.ContextHandle, fmt.Errorf("NetrShareDelStart failed: %s", srvsvc.StatusString(uint32(resp.Status)))
	}
	return resp.ContextHandle, nil
}
