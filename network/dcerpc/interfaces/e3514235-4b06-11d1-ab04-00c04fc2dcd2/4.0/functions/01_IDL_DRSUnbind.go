package functions

import (
	"fmt"

	drsuapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdrsr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-drsr"
)

// iDL_DRSUnbindRequest carries the [in] parameters of IDL_DRSUnbind.
type iDL_DRSUnbindRequest struct {
	PhDrs msdrsr.DRS_HANDLE
}

func (*iDL_DRSUnbindRequest) Opnum() uint16 { return drsuapi.OpnumIDL_DRSUnbind }

// iDL_DRSUnbindResponse carries the [out] parameters and return value of IDL_DRSUnbind.
type iDL_DRSUnbindResponse struct {
	PhDrs  msdrsr.DRS_HANDLE
	Status ndr.DWORD `ndr:"retval"`
}

// IDL_DRSUnbind calls IDL_DRSUnbind (opnum 1) ([MS-DRSR] — verify the parameter
// modeling and status handling).
func IDL_DRSUnbind(rpc ndr.Invoker, phDrs msdrsr.DRS_HANDLE) (PhDrs msdrsr.DRS_HANDLE, err error) {
	req := &iDL_DRSUnbindRequest{
		PhDrs: phDrs,
	}
	var resp iDL_DRSUnbindResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("IDL_DRSUnbind: %w", err)
		return
	}
	PhDrs = resp.PhDrs
	if uint32(resp.Status) != drsuapi.StatusSuccess {
		err = fmt.Errorf("IDL_DRSUnbind failed: %s", drsuapi.StatusString(uint32(resp.Status)))
	}
	return
}
