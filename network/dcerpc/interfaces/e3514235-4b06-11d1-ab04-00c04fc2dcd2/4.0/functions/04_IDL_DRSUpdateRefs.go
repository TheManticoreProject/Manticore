package functions

import (
	"fmt"

	drsuapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdrsr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-drsr"
)

// iDL_DRSUpdateRefsRequest carries the [in] parameters of IDL_DRSUpdateRefs.
type iDL_DRSUpdateRefsRequest struct {
	HDrs        msdrsr.DRS_HANDLE
	DwVersion   ndr.DWORD
	PmsgUpdRefs msdrsr.DRS_MSG_UPDREFS
}

func (*iDL_DRSUpdateRefsRequest) Opnum() uint16 { return drsuapi.OpnumIDL_DRSUpdateRefs }

// iDL_DRSUpdateRefsResponse carries the [out] parameters and return value of IDL_DRSUpdateRefs.
type iDL_DRSUpdateRefsResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// IDL_DRSUpdateRefs calls IDL_DRSUpdateRefs (opnum 4) ([MS-DRSR] — verify the parameter
// modeling and status handling).
func IDL_DRSUpdateRefs(rpc ndr.Invoker, hDrs msdrsr.DRS_HANDLE, dwVersion ndr.DWORD, pmsgUpdRefs msdrsr.DRS_MSG_UPDREFS) (err error) {
	req := &iDL_DRSUpdateRefsRequest{
		HDrs:        hDrs,
		DwVersion:   dwVersion,
		PmsgUpdRefs: pmsgUpdRefs,
	}
	var resp iDL_DRSUpdateRefsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("IDL_DRSUpdateRefs: %w", err)
		return
	}
	if uint32(resp.Status) != drsuapi.StatusSuccess {
		err = fmt.Errorf("IDL_DRSUpdateRefs failed: %s", drsuapi.StatusString(uint32(resp.Status)))
	}
	return
}
