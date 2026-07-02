package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// apiGetQuorumResourceRequest carries the [in] parameters of ApiGetQuorumResource.
type apiGetQuorumResourceRequest struct {
}

func (*apiGetQuorumResourceRequest) Opnum() uint16 { return clusapi.OpnumApiGetQuorumResource }

// apiGetQuorumResourceResponse carries the [out] parameters and return value of ApiGetQuorumResource.
type apiGetQuorumResourceResponse struct {
	LpszResourceName    ndr.WSTR
	LpszDeviceName      ndr.WSTR
	PdwMaxQuorumLogSize ndr.DWORD
	Rpc_status          ndr.DWORD
	Status              ndr.DWORD `ndr:"retval"`
}

// ApiGetQuorumResource calls ApiGetQuorumResource (opnum 5) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiGetQuorumResource(rpc ndr.Invoker) (LpszResourceName ndr.WSTR, LpszDeviceName ndr.WSTR, PdwMaxQuorumLogSize ndr.DWORD, Rpc_status ndr.DWORD, err error) {
	req := &apiGetQuorumResourceRequest{}
	var resp apiGetQuorumResourceResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiGetQuorumResource: %w", err)
		return
	}
	LpszResourceName = resp.LpszResourceName
	LpszDeviceName = resp.LpszDeviceName
	PdwMaxQuorumLogSize = resp.PdwMaxQuorumLogSize
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiGetQuorumResource failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
