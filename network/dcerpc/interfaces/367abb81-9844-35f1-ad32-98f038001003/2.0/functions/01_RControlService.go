package functions

import (
	"fmt"

	svcctl "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/367abb81-9844-35f1-ad32-98f038001003/2.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msscmr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-scmr"
)

// rControlServiceRequest carries the [in] parameters of RControlService.
type rControlServiceRequest struct {
	HService  msscmr.SC_RPC_HANDLE
	DwControl ndr.DWORD
}

func (*rControlServiceRequest) Opnum() uint16 { return svcctl.OpnumRControlService }

// rControlServiceResponse carries the [out] parameters and return value of RControlService.
type rControlServiceResponse struct {
	LpServiceStatus msscmr.SERVICE_STATUS
	Status          ndr.DWORD `ndr:"retval"`
}

// RControlService calls RControlService (opnum 1) ([MS-SCMR] — verify the parameter
// modeling and status handling).
func RControlService(rpc ndr.Invoker, hService msscmr.SC_RPC_HANDLE, dwControl ndr.DWORD) (LpServiceStatus msscmr.SERVICE_STATUS, err error) {
	req := &rControlServiceRequest{
		HService:  hService,
		DwControl: dwControl,
	}
	var resp rControlServiceResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RControlService: %w", err)
		return
	}
	LpServiceStatus = resp.LpServiceStatus
	if uint32(resp.Status) != svcctl.StatusSuccess {
		err = fmt.Errorf("RControlService failed: %s", svcctl.StatusString(uint32(resp.Status)))
	}
	return
}
