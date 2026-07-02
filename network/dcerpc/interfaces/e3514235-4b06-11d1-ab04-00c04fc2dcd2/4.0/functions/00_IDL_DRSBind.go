package functions

import (
	"fmt"

	drsuapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdrsr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-drsr"
)

// iDL_DRSBindRequest carries the [in] parameters of IDL_DRSBind.
type iDL_DRSBindRequest struct {
	PuuidClientDsa *msdrsr.UUID           `ndr:"unique"`
	PextClient     *msdrsr.DRS_EXTENSIONS `ndr:"unique"`
}

func (*iDL_DRSBindRequest) Opnum() uint16 { return drsuapi.OpnumIDL_DRSBind }

// iDL_DRSBindResponse carries the [out] parameters and return value of IDL_DRSBind.
// PpextServer models the IDL's [out] DRS_EXTENSIONS **ppextServer: the server-allocated
// extensions are returned behind a single unique referent.
type iDL_DRSBindResponse struct {
	PpextServer *msdrsr.DRS_EXTENSIONS `ndr:"unique"`
	PhDrs       msdrsr.DRS_HANDLE
	Status      ndr.DWORD `ndr:"retval"`
}

// IDL_DRSBind calls IDL_DRSBind (opnum 0) ([MS-DRSR] — verify the parameter
// modeling and status handling).
func IDL_DRSBind(rpc ndr.Invoker, puuidClientDsa *msdrsr.UUID, pextClient *msdrsr.DRS_EXTENSIONS) (PpextServer *msdrsr.DRS_EXTENSIONS, PhDrs msdrsr.DRS_HANDLE, err error) {
	req := &iDL_DRSBindRequest{
		PuuidClientDsa: puuidClientDsa,
		PextClient:     pextClient,
	}
	var resp iDL_DRSBindResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("IDL_DRSBind: %w", err)
		return
	}
	PpextServer = resp.PpextServer
	PhDrs = resp.PhDrs
	if uint32(resp.Status) != drsuapi.StatusSuccess {
		err = fmt.Errorf("IDL_DRSBind failed: %s", drsuapi.StatusString(uint32(resp.Status)))
	}
	return
}
