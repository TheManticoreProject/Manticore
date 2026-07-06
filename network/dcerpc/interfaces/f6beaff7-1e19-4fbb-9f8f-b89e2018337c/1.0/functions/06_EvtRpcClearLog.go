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

// evtRpcClearLogRequest carries the [in] parameters of EvtRpcClearLog.
type evtRpcClearLogRequest struct {
	Control     mseven6.PCONTEXT_HANDLE_OPERATION_CONTROL
	ChannelPath ndr.WSTR
	BackupPath  *ndr.WSTR `ndr:"unique"`
	Flags       ndr.DWORD
}

func (*evtRpcClearLogRequest) Opnum() uint16 { return IEventService.OpnumEvtRpcClearLog }

// evtRpcClearLogResponse carries the [out] parameters and return value of EvtRpcClearLog.
type evtRpcClearLogResponse struct {
	Error  mseven6.RpcInfo
	Status ndr.DWORD `ndr:"retval"`
}

// EvtRpcClearLog calls EvtRpcClearLog (opnum 6) ([MS-EVEN6] — verify the parameter
// modeling and status handling).
func EvtRpcClearLog(rpc ndr.Invoker, control mseven6.PCONTEXT_HANDLE_OPERATION_CONTROL, channelPath ndr.WSTR, backupPath *ndr.WSTR, flags ndr.DWORD) (Error mseven6.RpcInfo, err error) {
	req := &evtRpcClearLogRequest{
		Control:     control,
		ChannelPath: channelPath,
		BackupPath:  backupPath,
		Flags:       flags,
	}
	var resp evtRpcClearLogResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("EvtRpcClearLog: %w", err)
		return
	}
	Error = resp.Error
	if uint32(resp.Status) != IEventService.StatusSuccess {
		err = fmt.Errorf("EvtRpcClearLog failed: %s", IEventService.StatusString(uint32(resp.Status)))
	}
	return
}
