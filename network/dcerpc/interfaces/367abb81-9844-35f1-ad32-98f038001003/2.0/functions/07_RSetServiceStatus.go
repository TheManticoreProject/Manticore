package functions

import (
	"fmt"

	svcctl "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/367abb81-9844-35f1-ad32-98f038001003/2.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msscmr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-scmr"
)

// rSetServiceStatusRequest carries the [in] parameters of RSetServiceStatus.
type rSetServiceStatusRequest struct {
	HServiceStatus  msscmr.SC_RPC_HANDLE
	LpServiceStatus msscmr.SERVICE_STATUS
}

func (*rSetServiceStatusRequest) Opnum() uint16 { return svcctl.OpnumRSetServiceStatus }

// rSetServiceStatusResponse carries the [out] parameters and return value of RSetServiceStatus.
type rSetServiceStatusResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RSetServiceStatus calls RSetServiceStatus (opnum 7) ([MS-SCMR] — verify the parameter
// modeling and status handling).
func RSetServiceStatus(rpc ndr.Invoker, hServiceStatus msscmr.SC_RPC_HANDLE, lpServiceStatus msscmr.SERVICE_STATUS) (err error) {
	req := &rSetServiceStatusRequest{
		HServiceStatus:  hServiceStatus,
		LpServiceStatus: lpServiceStatus,
	}
	var resp rSetServiceStatusResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RSetServiceStatus: %w", err)
		return
	}
	if uint32(resp.Status) != svcctl.StatusSuccess {
		err = fmt.Errorf("RSetServiceStatus failed: %s", svcctl.StatusString(uint32(resp.Status)))
	}
	return
}
