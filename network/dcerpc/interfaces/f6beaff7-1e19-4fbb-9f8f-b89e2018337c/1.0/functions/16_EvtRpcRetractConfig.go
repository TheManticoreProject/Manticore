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

// evtRpcRetractConfigRequest carries the [in] parameters of EvtRpcRetractConfig.
type evtRpcRetractConfigRequest struct {
	Path  ndr.WSTR
	Flags ndr.DWORD
}

func (*evtRpcRetractConfigRequest) Opnum() uint16 { return IEventService.OpnumEvtRpcRetractConfig }

// evtRpcRetractConfigResponse carries the [out] parameters and return value of EvtRpcRetractConfig.
type evtRpcRetractConfigResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// EvtRpcRetractConfig calls EvtRpcRetractConfig (opnum 16) ([MS-EVEN6] — verify the parameter
// modeling and status handling).
func EvtRpcRetractConfig(rpc ndr.Invoker, path ndr.WSTR, flags ndr.DWORD) (err error) {
	req := &evtRpcRetractConfigRequest{
		Path:  path,
		Flags: flags,
	}
	var resp evtRpcRetractConfigResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("EvtRpcRetractConfig: %w", err)
		return
	}
	if uint32(resp.Status) != IEventService.StatusSuccess {
		err = fmt.Errorf("EvtRpcRetractConfig failed: %s", IEventService.StatusString(uint32(resp.Status)))
	}
	return
}
