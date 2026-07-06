package functions

import (
	"fmt"

	wkssvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6bffd098-a112-3610-9833-46c3f87e345a/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mswkst "github.com/TheManticoreProject/Manticore/windows/protocols/ms-wkst"
)

// netrWkstaSetInfoRequest carries the [in] parameters of NetrWkstaSetInfo.
type netrWkstaSetInfoRequest struct {
	ServerName     *ndr.WSTR `ndr:"unique"`
	Level          ndr.DWORD
	WkstaInfo      mswkst.WKSTA_INFO
	ErrorParameter *ndr.DWORD `ndr:"unique"`
}

func (*netrWkstaSetInfoRequest) Opnum() uint16 { return wkssvc.OpnumNetrWkstaSetInfo }

// netrWkstaSetInfoResponse carries the [out] parameters and return value of NetrWkstaSetInfo.
type netrWkstaSetInfoResponse struct {
	ErrorParameter *ndr.DWORD `ndr:"unique"`
	Status         ndr.DWORD  `ndr:"retval"`
}

// NetrWkstaSetInfo calls NetrWkstaSetInfo (opnum 1) ([MS-WKST] 3.2.4).
func NetrWkstaSetInfo(rpc ndr.Invoker, serverName *ndr.WSTR, level ndr.DWORD, wkstaInfo mswkst.WKSTA_INFO, errorParameter *ndr.DWORD) (ErrorParameter *ndr.DWORD, err error) {
	req := &netrWkstaSetInfoRequest{
		ServerName:     serverName,
		Level:          level,
		WkstaInfo:      wkstaInfo,
		ErrorParameter: errorParameter,
	}
	var resp netrWkstaSetInfoResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrWkstaSetInfo: %w", err)
		return
	}
	ErrorParameter = resp.ErrorParameter
	if uint32(resp.Status) != wkssvc.StatusSuccess {
		err = fmt.Errorf("NetrWkstaSetInfo failed: %s", wkssvc.StatusString(uint32(resp.Status)))
	}
	return
}
