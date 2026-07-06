package functions

import (
	"fmt"

	wkssvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6bffd098-a112-3610-9833-46c3f87e345a/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mswkst "github.com/TheManticoreProject/Manticore/windows/protocols/ms-wkst"
)

// netrWkstaTransportAddRequest carries the [in] parameters of NetrWkstaTransportAdd.
type netrWkstaTransportAddRequest struct {
	ServerName     *ndr.WSTR `ndr:"unique"`
	Level          ndr.DWORD
	TransportInfo  mswkst.WKSTA_TRANSPORT_INFO_0
	ErrorParameter *ndr.DWORD `ndr:"unique"`
}

func (*netrWkstaTransportAddRequest) Opnum() uint16 { return wkssvc.OpnumNetrWkstaTransportAdd }

// netrWkstaTransportAddResponse carries the [out] parameters and return value of NetrWkstaTransportAdd.
type netrWkstaTransportAddResponse struct {
	ErrorParameter *ndr.DWORD `ndr:"unique"`
	Status         ndr.DWORD  `ndr:"retval"`
}

// NetrWkstaTransportAdd calls NetrWkstaTransportAdd (opnum 6) ([MS-WKST] 3.2.4).
func NetrWkstaTransportAdd(rpc ndr.Invoker, serverName *ndr.WSTR, level ndr.DWORD, transportInfo mswkst.WKSTA_TRANSPORT_INFO_0, errorParameter *ndr.DWORD) (ErrorParameter *ndr.DWORD, err error) {
	req := &netrWkstaTransportAddRequest{
		ServerName:     serverName,
		Level:          level,
		TransportInfo:  transportInfo,
		ErrorParameter: errorParameter,
	}
	var resp netrWkstaTransportAddResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrWkstaTransportAdd: %w", err)
		return
	}
	ErrorParameter = resp.ErrorParameter
	if uint32(resp.Status) != wkssvc.StatusSuccess {
		err = fmt.Errorf("NetrWkstaTransportAdd failed: %s", wkssvc.StatusString(uint32(resp.Status)))
	}
	return
}
