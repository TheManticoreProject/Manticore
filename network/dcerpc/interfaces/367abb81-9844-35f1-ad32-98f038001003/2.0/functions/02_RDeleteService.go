package functions

import (
	"fmt"

	svcctl "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/367abb81-9844-35f1-ad32-98f038001003/2.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/367abb81-9844-35f1-ad32-98f038001003/2.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rDeleteServiceRequest carries the [in] parameters of RDeleteService.
type rDeleteServiceRequest struct {
	HService structures.SC_RPC_HANDLE
}

func (*rDeleteServiceRequest) Opnum() uint16 { return svcctl.OpnumRDeleteService }

// rDeleteServiceResponse carries the [out] parameters and return value of RDeleteService.
type rDeleteServiceResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RDeleteService calls RDeleteService (opnum 2) ([MS-SCMR] — verify the parameter
// modeling and status handling).
func RDeleteService(rpc ndr.Invoker, hService structures.SC_RPC_HANDLE) (err error) {
	req := &rDeleteServiceRequest{
		HService: hService,
	}
	var resp rDeleteServiceResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RDeleteService: %w", err)
		return
	}
	if uint32(resp.Status) != svcctl.StatusSuccess {
		err = fmt.Errorf("RDeleteService failed: %s", svcctl.StatusString(uint32(resp.Status)))
	}
	return
}
