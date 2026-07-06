package functions

// IDL source: [MS-CMRP] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cmrp/e12b6d8f-c410-49d7-a27d-9992782a9027
// A fetched copy is kept at ms-cmrp.idl in the interface directory.

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiCreateResTypeEnumRequest carries the [in] parameters of ApiCreateResTypeEnum.
type apiCreateResTypeEnumRequest struct {
	LpszTypeName ndr.WSTR
	DwType       ndr.DWORD
}

func (*apiCreateResTypeEnumRequest) Opnum() uint16 { return clusapi.OpnumApiCreateResTypeEnum }

// apiCreateResTypeEnumResponse carries the [out] parameters and return value of ApiCreateResTypeEnum.
type apiCreateResTypeEnumResponse struct {
	ReturnEnum *mscmrp.ENUM_LIST `ndr:"unique"`
	Rpc_status ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// ApiCreateResTypeEnum calls ApiCreateResTypeEnum (opnum 103) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiCreateResTypeEnum(rpc ndr.Invoker, lpszTypeName ndr.WSTR, dwType ndr.DWORD) (ReturnEnum *mscmrp.ENUM_LIST, Rpc_status ndr.DWORD, err error) {
	req := &apiCreateResTypeEnumRequest{
		LpszTypeName: lpszTypeName,
		DwType:       dwType,
	}
	var resp apiCreateResTypeEnumResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiCreateResTypeEnum: %w", err)
		return
	}
	ReturnEnum = resp.ReturnEnum
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiCreateResTypeEnum failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
