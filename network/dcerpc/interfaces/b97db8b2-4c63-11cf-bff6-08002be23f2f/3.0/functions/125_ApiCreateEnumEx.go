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

// apiCreateEnumExRequest carries the [in] parameters of ApiCreateEnumEx.
type apiCreateEnumExRequest struct {
	HCluster  mscmrp.HCLUSTER_RPC
	DwType    ndr.DWORD
	DwOptions ndr.DWORD
}

func (*apiCreateEnumExRequest) Opnum() uint16 { return clusapi.OpnumApiCreateEnumEx }

// apiCreateEnumExResponse carries the [out] parameters and return value of ApiCreateEnumEx.
type apiCreateEnumExResponse struct {
	ReturnIdEnum   *mscmrp.ENUM_LIST `ndr:"unique"`
	ReturnNameEnum *mscmrp.ENUM_LIST `ndr:"unique"`
	Rpc_status     ndr.DWORD
	Status         ndr.DWORD `ndr:"retval"`
}

// ApiCreateEnumEx calls ApiCreateEnumEx (opnum 125) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiCreateEnumEx(rpc ndr.Invoker, hCluster mscmrp.HCLUSTER_RPC, dwType ndr.DWORD, dwOptions ndr.DWORD) (ReturnIdEnum *mscmrp.ENUM_LIST, ReturnNameEnum *mscmrp.ENUM_LIST, Rpc_status ndr.DWORD, err error) {
	req := &apiCreateEnumExRequest{
		HCluster:  hCluster,
		DwType:    dwType,
		DwOptions: dwOptions,
	}
	var resp apiCreateEnumExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiCreateEnumEx: %w", err)
		return
	}
	ReturnIdEnum = resp.ReturnIdEnum
	ReturnNameEnum = resp.ReturnNameEnum
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiCreateEnumEx failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
