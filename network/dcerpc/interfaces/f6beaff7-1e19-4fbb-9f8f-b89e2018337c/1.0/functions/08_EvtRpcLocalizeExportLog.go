package functions

import (
	"fmt"

	IEventService "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/f6beaff7-1e19-4fbb-9f8f-b89e2018337c/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mseven6 "github.com/TheManticoreProject/Manticore/windows/protocols/ms-even6"
)

// evtRpcLocalizeExportLogRequest carries the [in] parameters of EvtRpcLocalizeExportLog.
type evtRpcLocalizeExportLogRequest struct {
	Control     mseven6.PCONTEXT_HANDLE_OPERATION_CONTROL
	LogFilePath ndr.WSTR
	Locale      ndr.DWORD
	Flags       ndr.DWORD
}

func (*evtRpcLocalizeExportLogRequest) Opnum() uint16 {
	return IEventService.OpnumEvtRpcLocalizeExportLog
}

// evtRpcLocalizeExportLogResponse carries the [out] parameters and return value of EvtRpcLocalizeExportLog.
type evtRpcLocalizeExportLogResponse struct {
	Error  mseven6.RpcInfo
	Status ndr.DWORD `ndr:"retval"`
}

// EvtRpcLocalizeExportLog calls EvtRpcLocalizeExportLog (opnum 8) ([MS-EVEN6] — verify the parameter
// modeling and status handling).
func EvtRpcLocalizeExportLog(rpc ndr.Invoker, control mseven6.PCONTEXT_HANDLE_OPERATION_CONTROL, logFilePath ndr.WSTR, locale ndr.DWORD, flags ndr.DWORD) (Error mseven6.RpcInfo, err error) {
	req := &evtRpcLocalizeExportLogRequest{
		Control:     control,
		LogFilePath: logFilePath,
		Locale:      locale,
		Flags:       flags,
	}
	var resp evtRpcLocalizeExportLogResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("EvtRpcLocalizeExportLog: %w", err)
		return
	}
	Error = resp.Error
	if uint32(resp.Status) != IEventService.StatusSuccess {
		err = fmt.Errorf("EvtRpcLocalizeExportLog failed: %s", IEventService.StatusString(uint32(resp.Status)))
	}
	return
}
