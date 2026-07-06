package functions

import (
	"fmt"

	wkssvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6bffd098-a112-3610-9833-46c3f87e345a/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mswkst "github.com/TheManticoreProject/Manticore/windows/protocols/ms-wkst"
)

// netrWkstaGetInfoRequest carries the [in] parameters of NetrWkstaGetInfo.
type netrWkstaGetInfoRequest struct {
	ServerName *ndr.WSTR `ndr:"unique"`
	Level      ndr.DWORD
}

func (*netrWkstaGetInfoRequest) Opnum() uint16 { return wkssvc.OpnumNetrWkstaGetInfo }

// netrWkstaGetInfoResponse carries the [out] parameters and return value of NetrWkstaGetInfo.
type netrWkstaGetInfoResponse struct {
	WkstaInfo mswkst.WKSTA_INFO
	Status    ndr.DWORD `ndr:"retval"`
}

// NetrWkstaGetInfo calls NetrWkstaGetInfo (opnum 0) ([MS-WKST] 3.2.4).
func NetrWkstaGetInfo(rpc ndr.Invoker, serverName *ndr.WSTR, level ndr.DWORD) (WkstaInfo mswkst.WKSTA_INFO, err error) {
	req := &netrWkstaGetInfoRequest{
		ServerName: serverName,
		Level:      level,
	}
	var resp netrWkstaGetInfoResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrWkstaGetInfo: %w", err)
		return
	}
	WkstaInfo = resp.WkstaInfo
	if uint32(resp.Status) != wkssvc.StatusSuccess {
		err = fmt.Errorf("NetrWkstaGetInfo failed: %s", wkssvc.StatusString(uint32(resp.Status)))
	}
	return
}
