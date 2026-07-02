package functions

import (
	"fmt"

	IEventService "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/f6beaff7-1e19-4fbb-9f8f-b89e2018337c/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mseven6 "github.com/TheManticoreProject/Manticore/windows/protocols/ms-even6"
)

// evtRpcGetChannelConfigRequest carries the [in] parameters of EvtRpcGetChannelConfig.
type evtRpcGetChannelConfigRequest struct {
	ChannelPath ndr.WSTR
	Flags       ndr.DWORD
}

func (*evtRpcGetChannelConfigRequest) Opnum() uint16 {
	return IEventService.OpnumEvtRpcGetChannelConfig
}

// evtRpcGetChannelConfigResponse carries the [out] parameters and return value of EvtRpcGetChannelConfig.
type evtRpcGetChannelConfigResponse struct {
	Props  mseven6.EvtRpcVariantList
	Status ndr.DWORD `ndr:"retval"`
}

// EvtRpcGetChannelConfig calls EvtRpcGetChannelConfig (opnum 20) ([MS-EVEN6] — verify the parameter
// modeling and status handling).
func EvtRpcGetChannelConfig(rpc ndr.Invoker, channelPath ndr.WSTR, flags ndr.DWORD) (Props mseven6.EvtRpcVariantList, err error) {
	req := &evtRpcGetChannelConfigRequest{
		ChannelPath: channelPath,
		Flags:       flags,
	}
	var resp evtRpcGetChannelConfigResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("EvtRpcGetChannelConfig: %w", err)
		return
	}
	Props = resp.Props
	if uint32(resp.Status) != IEventService.StatusSuccess {
		err = fmt.Errorf("EvtRpcGetChannelConfig failed: %s", IEventService.StatusString(uint32(resp.Status)))
	}
	return
}
