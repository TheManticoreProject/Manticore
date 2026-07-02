package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiCreateGroupResourceEnumRequest carries the [in] parameters of ApiCreateGroupResourceEnum.
type apiCreateGroupResourceEnumRequest struct {
	HGroup mscmrp.HGROUP_RPC
	DwType ndr.DWORD
}

func (*apiCreateGroupResourceEnumRequest) Opnum() uint16 {
	return clusapi.OpnumApiCreateGroupResourceEnum
}

// apiCreateGroupResourceEnumResponse carries the [out] parameters and return value of ApiCreateGroupResourceEnum.
type apiCreateGroupResourceEnumResponse struct {
	ReturnEnum *mscmrp.ENUM_LIST `ndr:"unique"`
	Rpc_status ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// ApiCreateGroupResourceEnum calls ApiCreateGroupResourceEnum (opnum 53) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiCreateGroupResourceEnum(rpc ndr.Invoker, hGroup mscmrp.HGROUP_RPC, dwType ndr.DWORD) (ReturnEnum *mscmrp.ENUM_LIST, Rpc_status ndr.DWORD, err error) {
	req := &apiCreateGroupResourceEnumRequest{
		HGroup: hGroup,
		DwType: dwType,
	}
	var resp apiCreateGroupResourceEnumResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiCreateGroupResourceEnum: %w", err)
		return
	}
	ReturnEnum = resp.ReturnEnum
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiCreateGroupResourceEnum failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
