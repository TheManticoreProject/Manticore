package functions

import (
	"fmt"

	svcctl "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/367abb81-9844-35f1-ad32-98f038001003/2.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msscmr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-scmr"
)

// rQueryServiceStatusRequest carries the [in] parameters of RQueryServiceStatus.
type rQueryServiceStatusRequest struct {
	HService msscmr.SC_RPC_HANDLE
}

func (*rQueryServiceStatusRequest) Opnum() uint16 { return svcctl.OpnumRQueryServiceStatus }

// rQueryServiceStatusResponse carries the [out] parameters and return value of RQueryServiceStatus.
type rQueryServiceStatusResponse struct {
	LpServiceStatus msscmr.SERVICE_STATUS
	Status          ndr.DWORD `ndr:"retval"`
}

// RQueryServiceStatus calls RQueryServiceStatus (opnum 6) ([MS-SCMR] — verify the parameter
// modeling and status handling).
func RQueryServiceStatus(rpc ndr.Invoker, hService msscmr.SC_RPC_HANDLE) (LpServiceStatus msscmr.SERVICE_STATUS, err error) {
	req := &rQueryServiceStatusRequest{
		HService: hService,
	}
	var resp rQueryServiceStatusResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RQueryServiceStatus: %w", err)
		return
	}
	LpServiceStatus = resp.LpServiceStatus
	if uint32(resp.Status) != svcctl.StatusSuccess {
		err = fmt.Errorf("RQueryServiceStatus failed: %s", svcctl.StatusString(uint32(resp.Status)))
	}
	return
}
