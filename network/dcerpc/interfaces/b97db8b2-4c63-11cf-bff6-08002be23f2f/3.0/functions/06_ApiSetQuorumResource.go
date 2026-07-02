package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiSetQuorumResourceRequest carries the [in] parameters of ApiSetQuorumResource.
type apiSetQuorumResourceRequest struct {
	HResource          mscmrp.HRES_RPC
	LpszDeviceName     ndr.WSTR
	DwMaxQuorumLogSize ndr.DWORD
}

func (*apiSetQuorumResourceRequest) Opnum() uint16 { return clusapi.OpnumApiSetQuorumResource }

// apiSetQuorumResourceResponse carries the [out] parameters and return value of ApiSetQuorumResource.
type apiSetQuorumResourceResponse struct {
	Rpc_status ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// ApiSetQuorumResource calls ApiSetQuorumResource (opnum 6) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiSetQuorumResource(rpc ndr.Invoker, hResource mscmrp.HRES_RPC, lpszDeviceName ndr.WSTR, dwMaxQuorumLogSize ndr.DWORD) (Rpc_status ndr.DWORD, err error) {
	req := &apiSetQuorumResourceRequest{
		HResource:          hResource,
		LpszDeviceName:     lpszDeviceName,
		DwMaxQuorumLogSize: dwMaxQuorumLogSize,
	}
	var resp apiSetQuorumResourceResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiSetQuorumResource: %w", err)
		return
	}
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiSetQuorumResource failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
