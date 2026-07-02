package functions

import (
	"fmt"

	IEventService "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/f6beaff7-1e19-4fbb-9f8f-b89e2018337c/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// evtRpcGetPublisherListRequest carries the [in] parameters of EvtRpcGetPublisherList.
type evtRpcGetPublisherListRequest struct {
	Flags ndr.DWORD
}

func (*evtRpcGetPublisherListRequest) Opnum() uint16 {
	return IEventService.OpnumEvtRpcGetPublisherList
}

// evtRpcGetPublisherListResponse carries the [out] parameters and return value of EvtRpcGetPublisherList.
type evtRpcGetPublisherListResponse struct {
	NumPublisherIds ndr.DWORD
	PublisherIds    []*ndr.WSTR `ndr:"unique,size_is=NumPublisherIds,elem=unique"`
	Status          ndr.DWORD   `ndr:"retval"`
}

// EvtRpcGetPublisherList calls EvtRpcGetPublisherList (opnum 22) ([MS-EVEN6] — verify the parameter
// modeling and status handling).
func EvtRpcGetPublisherList(rpc ndr.Invoker, flags ndr.DWORD) (NumPublisherIds ndr.DWORD, PublisherIds []*ndr.WSTR, err error) {
	req := &evtRpcGetPublisherListRequest{
		Flags: flags,
	}
	var resp evtRpcGetPublisherListResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("EvtRpcGetPublisherList: %w", err)
		return
	}
	NumPublisherIds = resp.NumPublisherIds
	PublisherIds = resp.PublisherIds
	if uint32(resp.Status) != IEventService.StatusSuccess {
		err = fmt.Errorf("EvtRpcGetPublisherList failed: %s", IEventService.StatusString(uint32(resp.Status)))
	}
	return
}
