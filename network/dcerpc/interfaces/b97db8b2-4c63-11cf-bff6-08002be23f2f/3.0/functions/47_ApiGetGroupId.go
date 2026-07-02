package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiGetGroupIdRequest carries the [in] parameters of ApiGetGroupId.
type apiGetGroupIdRequest struct {
	HGroup mscmrp.HGROUP_RPC
}

func (*apiGetGroupIdRequest) Opnum() uint16 { return clusapi.OpnumApiGetGroupId }

// apiGetGroupIdResponse carries the [out] parameters and return value of ApiGetGroupId.
type apiGetGroupIdResponse struct {
	PGuid      ndr.WSTR
	Rpc_status ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// ApiGetGroupId calls ApiGetGroupId (opnum 47) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiGetGroupId(rpc ndr.Invoker, hGroup mscmrp.HGROUP_RPC) (PGuid ndr.WSTR, Rpc_status ndr.DWORD, err error) {
	req := &apiGetGroupIdRequest{
		HGroup: hGroup,
	}
	var resp apiGetGroupIdResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiGetGroupId: %w", err)
		return
	}
	PGuid = resp.PGuid
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiGetGroupId failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
