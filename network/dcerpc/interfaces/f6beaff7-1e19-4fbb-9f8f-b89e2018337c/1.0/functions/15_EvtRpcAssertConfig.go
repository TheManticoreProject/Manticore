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

// evtRpcAssertConfigRequest carries the [in] parameters of EvtRpcAssertConfig.
type evtRpcAssertConfigRequest struct {
	Path  ndr.WSTR
	Flags ndr.DWORD
}

func (*evtRpcAssertConfigRequest) Opnum() uint16 { return IEventService.OpnumEvtRpcAssertConfig }

// evtRpcAssertConfigResponse carries the [out] parameters and return value of EvtRpcAssertConfig.
type evtRpcAssertConfigResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// EvtRpcAssertConfig calls EvtRpcAssertConfig (opnum 15) ([MS-EVEN6] — verify the parameter
// modeling and status handling).
func EvtRpcAssertConfig(rpc ndr.Invoker, path ndr.WSTR, flags ndr.DWORD) (err error) {
	req := &evtRpcAssertConfigRequest{
		Path:  path,
		Flags: flags,
	}
	var resp evtRpcAssertConfigResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("EvtRpcAssertConfig: %w", err)
		return
	}
	if uint32(resp.Status) != IEventService.StatusSuccess {
		err = fmt.Errorf("EvtRpcAssertConfig failed: %s", IEventService.StatusString(uint32(resp.Status)))
	}
	return
}
