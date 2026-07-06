package functions

// IDL source: [MS-EVEN6] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-even6/2d808edd-719a-4c69-b34a-df766adb5f0c
// A fetched copy is kept at ms-even6.idl in the interface directory.

import (
	"fmt"

	IEventService "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/f6beaff7-1e19-4fbb-9f8f-b89e2018337c/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// evtRpcGetPublisherListForChannelRequest carries the [in] parameters of EvtRpcGetPublisherListForChannel.
type evtRpcGetPublisherListForChannelRequest struct {
	ChannelName ndr.WSTR
	Flags       ndr.DWORD
}

func (*evtRpcGetPublisherListForChannelRequest) Opnum() uint16 {
	return IEventService.OpnumEvtRpcGetPublisherListForChannel
}

// evtRpcGetPublisherListForChannelResponse carries the [out] parameters and return value of EvtRpcGetPublisherListForChannel.
type evtRpcGetPublisherListForChannelResponse struct {
	NumPublisherIds ndr.DWORD
	PublisherIds    []*ndr.WSTR `ndr:"unique,size_is=NumPublisherIds,elem=unique"`
	Status          ndr.DWORD   `ndr:"retval"`
}

// EvtRpcGetPublisherListForChannel calls EvtRpcGetPublisherListForChannel (opnum 23) ([MS-EVEN6] — verify the parameter
// modeling and status handling).
func EvtRpcGetPublisherListForChannel(rpc ndr.Invoker, channelName ndr.WSTR, flags ndr.DWORD) (NumPublisherIds ndr.DWORD, PublisherIds []*ndr.WSTR, err error) {
	req := &evtRpcGetPublisherListForChannelRequest{
		ChannelName: channelName,
		Flags:       flags,
	}
	var resp evtRpcGetPublisherListForChannelResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("EvtRpcGetPublisherListForChannel: %w", err)
		return
	}
	NumPublisherIds = resp.NumPublisherIds
	PublisherIds = resp.PublisherIds
	if uint32(resp.Status) != IEventService.StatusSuccess {
		err = fmt.Errorf("EvtRpcGetPublisherListForChannel failed: %s", IEventService.StatusString(uint32(resp.Status)))
	}
	return
}
