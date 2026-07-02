package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiCreateEnumRequest carries the [in] parameters of ApiCreateEnum.
type apiCreateEnumRequest struct {
	DwType ndr.DWORD
}

func (*apiCreateEnumRequest) Opnum() uint16 { return clusapi.OpnumApiCreateEnum }

// apiCreateEnumResponse carries the [out] parameters and return value of ApiCreateEnum.
type apiCreateEnumResponse struct {
	ReturnEnum *mscmrp.ENUM_LIST `ndr:"unique"`
	Rpc_status ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// ApiCreateEnum calls ApiCreateEnum (opnum 7) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiCreateEnum(rpc ndr.Invoker, dwType ndr.DWORD) (ReturnEnum *mscmrp.ENUM_LIST, Rpc_status ndr.DWORD, err error) {
	req := &apiCreateEnumRequest{
		DwType: dwType,
	}
	var resp apiCreateEnumResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiCreateEnum: %w", err)
		return
	}
	ReturnEnum = resp.ReturnEnum
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiCreateEnum failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
