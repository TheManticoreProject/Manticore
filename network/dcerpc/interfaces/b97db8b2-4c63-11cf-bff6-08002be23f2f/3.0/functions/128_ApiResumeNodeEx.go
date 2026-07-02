package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiResumeNodeExRequest carries the [in] parameters of ApiResumeNodeEx.
type apiResumeNodeExRequest struct {
	HNode                 mscmrp.HNODE_RPC
	DwResumeFailbackType  ndr.DWORD
	DwResumeFlagsReserved ndr.DWORD
}

func (*apiResumeNodeExRequest) Opnum() uint16 { return clusapi.OpnumApiResumeNodeEx }

// apiResumeNodeExResponse carries the [out] parameters and return value of ApiResumeNodeEx.
type apiResumeNodeExResponse struct {
	Rpc_status ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// ApiResumeNodeEx calls ApiResumeNodeEx (opnum 128) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiResumeNodeEx(rpc ndr.Invoker, hNode mscmrp.HNODE_RPC, dwResumeFailbackType ndr.DWORD, dwResumeFlagsReserved ndr.DWORD) (Rpc_status ndr.DWORD, err error) {
	req := &apiResumeNodeExRequest{
		HNode:                 hNode,
		DwResumeFailbackType:  dwResumeFailbackType,
		DwResumeFlagsReserved: dwResumeFlagsReserved,
	}
	var resp apiResumeNodeExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiResumeNodeEx: %w", err)
		return
	}
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiResumeNodeEx failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
