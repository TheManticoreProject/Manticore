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

// apiCreateNodeEnumRequest carries the [in] parameters of ApiCreateNodeEnum.
type apiCreateNodeEnumRequest struct {
	HNode  mscmrp.HNODE_RPC
	DwType ndr.DWORD
}

func (*apiCreateNodeEnumRequest) Opnum() uint16 { return clusapi.OpnumApiCreateNodeEnum }

// apiCreateNodeEnumResponse carries the [out] parameters and return value of ApiCreateNodeEnum.
type apiCreateNodeEnumResponse struct {
	ReturnEnum *mscmrp.ENUM_LIST `ndr:"unique"`
	Rpc_status ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// ApiCreateNodeEnum calls ApiCreateNodeEnum (opnum 101) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiCreateNodeEnum(rpc ndr.Invoker, hNode mscmrp.HNODE_RPC, dwType ndr.DWORD) (ReturnEnum *mscmrp.ENUM_LIST, Rpc_status ndr.DWORD, err error) {
	req := &apiCreateNodeEnumRequest{
		HNode:  hNode,
		DwType: dwType,
	}
	var resp apiCreateNodeEnumResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiCreateNodeEnum: %w", err)
		return
	}
	ReturnEnum = resp.ReturnEnum
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiCreateNodeEnum failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
