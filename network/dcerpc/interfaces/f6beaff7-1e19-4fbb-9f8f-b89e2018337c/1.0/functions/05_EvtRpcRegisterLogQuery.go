package functions

import (
	"fmt"

	IEventService "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/f6beaff7-1e19-4fbb-9f8f-b89e2018337c/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mseven6 "github.com/TheManticoreProject/Manticore/windows/protocols/ms-even6"
)

// evtRpcRegisterLogQueryRequest carries the [in] parameters of EvtRpcRegisterLogQuery.
type evtRpcRegisterLogQueryRequest struct {
	Path  *ndr.WSTR `ndr:"unique"`
	Query ndr.WSTR
	Flags ndr.DWORD
}

func (*evtRpcRegisterLogQueryRequest) Opnum() uint16 {
	return IEventService.OpnumEvtRpcRegisterLogQuery
}

// evtRpcRegisterLogQueryResponse carries the [out] parameters and return value of EvtRpcRegisterLogQuery.
type evtRpcRegisterLogQueryResponse struct {
	Handle               mseven6.PCONTEXT_HANDLE_LOG_QUERY
	OpControl            mseven6.PCONTEXT_HANDLE_OPERATION_CONTROL
	QueryChannelInfoSize ndr.DWORD
	QueryChannelInfo     []mseven6.EvtRpcQueryChannelInfo `ndr:"unique,size_is=QueryChannelInfoSize"`
	Error                mseven6.RpcInfo
	Status               ndr.DWORD `ndr:"retval"`
}

// EvtRpcRegisterLogQuery calls EvtRpcRegisterLogQuery (opnum 5) ([MS-EVEN6] — verify the parameter
// modeling and status handling).
func EvtRpcRegisterLogQuery(rpc ndr.Invoker, path *ndr.WSTR, query ndr.WSTR, flags ndr.DWORD) (Handle mseven6.PCONTEXT_HANDLE_LOG_QUERY, OpControl mseven6.PCONTEXT_HANDLE_OPERATION_CONTROL, QueryChannelInfoSize ndr.DWORD, QueryChannelInfo []mseven6.EvtRpcQueryChannelInfo, Error mseven6.RpcInfo, err error) {
	req := &evtRpcRegisterLogQueryRequest{
		Path:  path,
		Query: query,
		Flags: flags,
	}
	var resp evtRpcRegisterLogQueryResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("EvtRpcRegisterLogQuery: %w", err)
		return
	}
	Handle = resp.Handle
	OpControl = resp.OpControl
	QueryChannelInfoSize = resp.QueryChannelInfoSize
	QueryChannelInfo = resp.QueryChannelInfo
	Error = resp.Error
	if uint32(resp.Status) != IEventService.StatusSuccess {
		err = fmt.Errorf("EvtRpcRegisterLogQuery failed: %s", IEventService.StatusString(uint32(resp.Status)))
	}
	return
}
