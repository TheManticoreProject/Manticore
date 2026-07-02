package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiCreateResEnumRequest carries the [in] parameters of ApiCreateResEnum.
type apiCreateResEnumRequest struct {
	HResource mscmrp.HRES_RPC
	DwType    ndr.DWORD
}

func (*apiCreateResEnumRequest) Opnum() uint16 { return clusapi.OpnumApiCreateResEnum }

// apiCreateResEnumResponse carries the [out] parameters and return value of ApiCreateResEnum.
type apiCreateResEnumResponse struct {
	ReturnEnum *mscmrp.ENUM_LIST `ndr:"unique"`
	Rpc_status ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// ApiCreateResEnum calls ApiCreateResEnum (opnum 22) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiCreateResEnum(rpc ndr.Invoker, hResource mscmrp.HRES_RPC, dwType ndr.DWORD) (ReturnEnum *mscmrp.ENUM_LIST, Rpc_status ndr.DWORD, err error) {
	req := &apiCreateResEnumRequest{
		HResource: hResource,
		DwType:    dwType,
	}
	var resp apiCreateResEnumResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiCreateResEnum: %w", err)
		return
	}
	ReturnEnum = resp.ReturnEnum
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiCreateResEnum failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
