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

// evtRpcGetChannelListRequest carries the [in] parameters of EvtRpcGetChannelList.
type evtRpcGetChannelListRequest struct {
	Flags ndr.DWORD
}

func (*evtRpcGetChannelListRequest) Opnum() uint16 { return IEventService.OpnumEvtRpcGetChannelList }

// evtRpcGetChannelListResponse carries the [out] parameters and return value of EvtRpcGetChannelList.
type evtRpcGetChannelListResponse struct {
	NumChannelPaths ndr.DWORD
	ChannelPaths    []*ndr.WSTR `ndr:"unique,size_is=NumChannelPaths,elem=unique"`
	Status          ndr.DWORD   `ndr:"retval"`
}

// EvtRpcGetChannelList calls EvtRpcGetChannelList (opnum 19) ([MS-EVEN6] — verify the parameter
// modeling and status handling).
func EvtRpcGetChannelList(rpc ndr.Invoker, flags ndr.DWORD) (NumChannelPaths ndr.DWORD, ChannelPaths []*ndr.WSTR, err error) {
	req := &evtRpcGetChannelListRequest{
		Flags: flags,
	}
	var resp evtRpcGetChannelListResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("EvtRpcGetChannelList: %w", err)
		return
	}
	NumChannelPaths = resp.NumChannelPaths
	ChannelPaths = resp.ChannelPaths
	if uint32(resp.Status) != IEventService.StatusSuccess {
		err = fmt.Errorf("EvtRpcGetChannelList failed: %s", IEventService.StatusString(uint32(resp.Status)))
	}
	return
}
