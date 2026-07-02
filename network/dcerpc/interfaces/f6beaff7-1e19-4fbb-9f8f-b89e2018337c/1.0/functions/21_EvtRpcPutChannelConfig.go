package functions

import (
	"fmt"

	IEventService "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/f6beaff7-1e19-4fbb-9f8f-b89e2018337c/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mseven6 "github.com/TheManticoreProject/Manticore/windows/protocols/ms-even6"
)

// evtRpcPutChannelConfigRequest carries the [in] parameters of EvtRpcPutChannelConfig.
type evtRpcPutChannelConfigRequest struct {
	ChannelPath ndr.WSTR
	Flags       ndr.DWORD
	Props       mseven6.EvtRpcVariantList
}

func (*evtRpcPutChannelConfigRequest) Opnum() uint16 {
	return IEventService.OpnumEvtRpcPutChannelConfig
}

// evtRpcPutChannelConfigResponse carries the [out] parameters and return value of EvtRpcPutChannelConfig.
type evtRpcPutChannelConfigResponse struct {
	Error  mseven6.RpcInfo
	Status ndr.DWORD `ndr:"retval"`
}

// EvtRpcPutChannelConfig calls EvtRpcPutChannelConfig (opnum 21) ([MS-EVEN6] — verify the parameter
// modeling and status handling).
func EvtRpcPutChannelConfig(rpc ndr.Invoker, channelPath ndr.WSTR, flags ndr.DWORD, props mseven6.EvtRpcVariantList) (Error mseven6.RpcInfo, err error) {
	req := &evtRpcPutChannelConfigRequest{
		ChannelPath: channelPath,
		Flags:       flags,
		Props:       props,
	}
	var resp evtRpcPutChannelConfigResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("EvtRpcPutChannelConfig: %w", err)
		return
	}
	Error = resp.Error
	if uint32(resp.Status) != IEventService.StatusSuccess {
		err = fmt.Errorf("EvtRpcPutChannelConfig failed: %s", IEventService.StatusString(uint32(resp.Status)))
	}
	return
}
