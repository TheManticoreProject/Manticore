package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiCreateNodeEnumExRequest carries the [in] parameters of ApiCreateNodeEnumEx.
type apiCreateNodeEnumExRequest struct {
	HNode     mscmrp.HNODE_RPC
	DwType    ndr.DWORD
	DwOptions ndr.DWORD
}

func (*apiCreateNodeEnumExRequest) Opnum() uint16 { return clusapi.OpnumApiCreateNodeEnumEx }

// apiCreateNodeEnumExResponse carries the [out] parameters and return value of ApiCreateNodeEnumEx.
type apiCreateNodeEnumExResponse struct {
	ReturnIdEnum   *mscmrp.ENUM_LIST `ndr:"unique"`
	ReturnNameEnum *mscmrp.ENUM_LIST `ndr:"unique"`
	Rpc_status     ndr.DWORD
	Status         ndr.DWORD `ndr:"retval"`
}

// ApiCreateNodeEnumEx calls ApiCreateNodeEnumEx (opnum 124) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiCreateNodeEnumEx(rpc ndr.Invoker, hNode mscmrp.HNODE_RPC, dwType ndr.DWORD, dwOptions ndr.DWORD) (ReturnIdEnum *mscmrp.ENUM_LIST, ReturnNameEnum *mscmrp.ENUM_LIST, Rpc_status ndr.DWORD, err error) {
	req := &apiCreateNodeEnumExRequest{
		HNode:     hNode,
		DwType:    dwType,
		DwOptions: dwOptions,
	}
	var resp apiCreateNodeEnumExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiCreateNodeEnumEx: %w", err)
		return
	}
	ReturnIdEnum = resp.ReturnIdEnum
	ReturnNameEnum = resp.ReturnNameEnum
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiCreateNodeEnumEx failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
