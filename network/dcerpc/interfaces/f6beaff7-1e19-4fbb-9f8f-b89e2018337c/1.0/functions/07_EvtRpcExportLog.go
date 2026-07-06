package functions

// IDL source: [MS-EVEN6] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-even6/2d808edd-719a-4c69-b34a-df766adb5f0c
// A fetched copy is kept at ms-even6.idl in the interface directory.

import (
	"fmt"

	IEventService "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/f6beaff7-1e19-4fbb-9f8f-b89e2018337c/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mseven6 "github.com/TheManticoreProject/Manticore/windows/protocols/ms-even6"
)

// evtRpcExportLogRequest carries the [in] parameters of EvtRpcExportLog.
type evtRpcExportLogRequest struct {
	Control     mseven6.PCONTEXT_HANDLE_OPERATION_CONTROL
	ChannelPath *ndr.WSTR `ndr:"unique"`
	Query       ndr.WSTR
	BackupPath  ndr.WSTR
	Flags       ndr.DWORD
}

func (*evtRpcExportLogRequest) Opnum() uint16 { return IEventService.OpnumEvtRpcExportLog }

// evtRpcExportLogResponse carries the [out] parameters and return value of EvtRpcExportLog.
type evtRpcExportLogResponse struct {
	Error  mseven6.RpcInfo
	Status ndr.DWORD `ndr:"retval"`
}

// EvtRpcExportLog calls EvtRpcExportLog (opnum 7) ([MS-EVEN6] — verify the parameter
// modeling and status handling).
func EvtRpcExportLog(rpc ndr.Invoker, control mseven6.PCONTEXT_HANDLE_OPERATION_CONTROL, channelPath *ndr.WSTR, query ndr.WSTR, backupPath ndr.WSTR, flags ndr.DWORD) (Error mseven6.RpcInfo, err error) {
	req := &evtRpcExportLogRequest{
		Control:     control,
		ChannelPath: channelPath,
		Query:       query,
		BackupPath:  backupPath,
		Flags:       flags,
	}
	var resp evtRpcExportLogResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("EvtRpcExportLog: %w", err)
		return
	}
	Error = resp.Error
	if uint32(resp.Status) != IEventService.StatusSuccess {
		err = fmt.Errorf("EvtRpcExportLog failed: %s", IEventService.StatusString(uint32(resp.Status)))
	}
	return
}
