package functions

import (
	"fmt"

	svcctl "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/367abb81-9844-35f1-ad32-98f038001003/2.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/367abb81-9844-35f1-ad32-98f038001003/2.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rGetNotifyResultsRequest carries the [in] parameters of RGetNotifyResults.
type rGetNotifyResultsRequest struct {
	HNotify structures.SC_NOTIFY_RPC_HANDLE
}

func (*rGetNotifyResultsRequest) Opnum() uint16 { return svcctl.OpnumRGetNotifyResults }

// rGetNotifyResultsResponse carries the [out] parameters and return value of RGetNotifyResults.
type rGetNotifyResultsResponse struct {
	PpNotifyParams *structures.SC_RPC_NOTIFY_PARAMS_LIST `ndr:"unique"`
	Status         ndr.DWORD                             `ndr:"retval"`
}

// RGetNotifyResults calls RGetNotifyResults (opnum 48) ([MS-SCMR] — verify the parameter
// modeling and status handling).
func RGetNotifyResults(rpc ndr.Invoker, hNotify structures.SC_NOTIFY_RPC_HANDLE) (PpNotifyParams *structures.SC_RPC_NOTIFY_PARAMS_LIST, err error) {
	req := &rGetNotifyResultsRequest{
		HNotify: hNotify,
	}
	var resp rGetNotifyResultsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RGetNotifyResults: %w", err)
		return
	}
	PpNotifyParams = resp.PpNotifyParams
	if uint32(resp.Status) != svcctl.StatusSuccess {
		err = fmt.Errorf("RGetNotifyResults failed: %s", svcctl.StatusString(uint32(resp.Status)))
	}
	return
}
