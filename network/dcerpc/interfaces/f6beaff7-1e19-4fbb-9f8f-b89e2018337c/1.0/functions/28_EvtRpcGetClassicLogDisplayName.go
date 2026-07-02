package functions

import (
	"fmt"

	IEventService "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/f6beaff7-1e19-4fbb-9f8f-b89e2018337c/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// evtRpcGetClassicLogDisplayNameRequest carries the [in] parameters of EvtRpcGetClassicLogDisplayName.
type evtRpcGetClassicLogDisplayNameRequest struct {
	LogName ndr.WSTR
	Locale  ndr.DWORD
	Flags   ndr.DWORD
}

func (*evtRpcGetClassicLogDisplayNameRequest) Opnum() uint16 {
	return IEventService.OpnumEvtRpcGetClassicLogDisplayName
}

// evtRpcGetClassicLogDisplayNameResponse carries the [out] parameters and return value of EvtRpcGetClassicLogDisplayName.
type evtRpcGetClassicLogDisplayNameResponse struct {
	DisplayName *ndr.WSTR `ndr:"unique"`
	Status      ndr.DWORD `ndr:"retval"`
}

// EvtRpcGetClassicLogDisplayName calls EvtRpcGetClassicLogDisplayName (opnum 28) ([MS-EVEN6] — verify the parameter
// modeling and status handling).
func EvtRpcGetClassicLogDisplayName(rpc ndr.Invoker, logName ndr.WSTR, locale ndr.DWORD, flags ndr.DWORD) (DisplayName *ndr.WSTR, err error) {
	req := &evtRpcGetClassicLogDisplayNameRequest{
		LogName: logName,
		Locale:  locale,
		Flags:   flags,
	}
	var resp evtRpcGetClassicLogDisplayNameResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("EvtRpcGetClassicLogDisplayName: %w", err)
		return
	}
	DisplayName = resp.DisplayName
	if uint32(resp.Status) != IEventService.StatusSuccess {
		err = fmt.Errorf("EvtRpcGetClassicLogDisplayName failed: %s", IEventService.StatusString(uint32(resp.Status)))
	}
	return
}
