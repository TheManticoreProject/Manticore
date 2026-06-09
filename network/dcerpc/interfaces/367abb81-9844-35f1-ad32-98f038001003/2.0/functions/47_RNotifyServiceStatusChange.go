package functions

import (
	"fmt"

	svcctl "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/367abb81-9844-35f1-ad32-98f038001003/2.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/367abb81-9844-35f1-ad32-98f038001003/2.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// rNotifyServiceStatusChangeRequest carries the [in] parameters of RNotifyServiceStatusChange.
type rNotifyServiceStatusChangeRequest struct {
	HService           structures.SC_RPC_HANDLE
	NotifyParams       structures.SC_RPC_NOTIFY_PARAMS
	PClientProcessGuid guid.GUID
}

func (*rNotifyServiceStatusChangeRequest) Opnum() uint16 {
	return svcctl.OpnumRNotifyServiceStatusChange
}

// rNotifyServiceStatusChangeResponse carries the [out] parameters and return value of RNotifyServiceStatusChange.
type rNotifyServiceStatusChangeResponse struct {
	PSCMProcessGuid     guid.GUID
	PfCreateRemoteQueue ndr.BOOL
	PhNotify            structures.LPSC_NOTIFY_RPC_HANDLE
	Status              ndr.DWORD `ndr:"retval"`
}

// RNotifyServiceStatusChange calls RNotifyServiceStatusChange (opnum 47) ([MS-SCMR] — verify the parameter
// modeling and status handling).
func RNotifyServiceStatusChange(rpc ndr.Invoker, hService structures.SC_RPC_HANDLE, notifyParams structures.SC_RPC_NOTIFY_PARAMS, pClientProcessGuid guid.GUID) (PSCMProcessGuid guid.GUID, PfCreateRemoteQueue ndr.BOOL, PhNotify structures.LPSC_NOTIFY_RPC_HANDLE, err error) {
	req := &rNotifyServiceStatusChangeRequest{
		HService:           hService,
		NotifyParams:       notifyParams,
		PClientProcessGuid: pClientProcessGuid,
	}
	var resp rNotifyServiceStatusChangeResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RNotifyServiceStatusChange: %w", err)
		return
	}
	PSCMProcessGuid = resp.PSCMProcessGuid
	PfCreateRemoteQueue = resp.PfCreateRemoteQueue
	PhNotify = resp.PhNotify
	if uint32(resp.Status) != svcctl.StatusSuccess {
		err = fmt.Errorf("RNotifyServiceStatusChange failed: %s", svcctl.StatusString(uint32(resp.Status)))
	}
	return
}
