package functions

import (
	"fmt"

	IEventService "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/f6beaff7-1e19-4fbb-9f8f-b89e2018337c/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mseven6 "github.com/TheManticoreProject/Manticore/windows/protocols/ms-even6"
)

// evtRpcRegisterRemoteSubscriptionRequest carries the [in] parameters of EvtRpcRegisterRemoteSubscription.
type evtRpcRegisterRemoteSubscriptionRequest struct {
	ChannelPath *ndr.WSTR `ndr:"unique"`
	Query       ndr.WSTR
	BookmarkXml *ndr.WSTR `ndr:"unique"`
	Flags       ndr.DWORD
}

func (*evtRpcRegisterRemoteSubscriptionRequest) Opnum() uint16 {
	return IEventService.OpnumEvtRpcRegisterRemoteSubscription
}

// evtRpcRegisterRemoteSubscriptionResponse carries the [out] parameters and return value of EvtRpcRegisterRemoteSubscription.
type evtRpcRegisterRemoteSubscriptionResponse struct {
	Handle               mseven6.PCONTEXT_HANDLE_REMOTE_SUBSCRIPTION
	Control              mseven6.PCONTEXT_HANDLE_OPERATION_CONTROL
	QueryChannelInfoSize ndr.DWORD
	QueryChannelInfo     []mseven6.EvtRpcQueryChannelInfo `ndr:"unique,size_is=QueryChannelInfoSize"`
	Error                mseven6.RpcInfo
	Status               ndr.DWORD `ndr:"retval"`
}

// EvtRpcRegisterRemoteSubscription calls EvtRpcRegisterRemoteSubscription (opnum 0) ([MS-EVEN6] — verify the parameter
// modeling and status handling).
func EvtRpcRegisterRemoteSubscription(rpc ndr.Invoker, channelPath *ndr.WSTR, query ndr.WSTR, bookmarkXml *ndr.WSTR, flags ndr.DWORD) (Handle mseven6.PCONTEXT_HANDLE_REMOTE_SUBSCRIPTION, Control mseven6.PCONTEXT_HANDLE_OPERATION_CONTROL, QueryChannelInfoSize ndr.DWORD, QueryChannelInfo []mseven6.EvtRpcQueryChannelInfo, Error mseven6.RpcInfo, err error) {
	req := &evtRpcRegisterRemoteSubscriptionRequest{
		ChannelPath: channelPath,
		Query:       query,
		BookmarkXml: bookmarkXml,
		Flags:       flags,
	}
	var resp evtRpcRegisterRemoteSubscriptionResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("EvtRpcRegisterRemoteSubscription: %w", err)
		return
	}
	Handle = resp.Handle
	Control = resp.Control
	QueryChannelInfoSize = resp.QueryChannelInfoSize
	QueryChannelInfo = resp.QueryChannelInfo
	Error = resp.Error
	if uint32(resp.Status) != IEventService.StatusSuccess {
		err = fmt.Errorf("EvtRpcRegisterRemoteSubscription failed: %s", IEventService.StatusString(uint32(resp.Status)))
	}
	return
}
