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

// apiSetGroupNodeListRequest carries the [in] parameters of ApiSetGroupNodeList.
type apiSetGroupNodeListRequest struct {
	HGroup          mscmrp.HGROUP_RPC
	MultiSzNodeList []uint16 `ndr:"ref,size_is=CchListSize"`
	CchListSize     ndr.DWORD
}

func (*apiSetGroupNodeListRequest) Opnum() uint16 { return clusapi.OpnumApiSetGroupNodeList }

// apiSetGroupNodeListResponse carries the [out] parameters and return value of ApiSetGroupNodeList.
type apiSetGroupNodeListResponse struct {
	Rpc_status ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// ApiSetGroupNodeList calls ApiSetGroupNodeList (opnum 54) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiSetGroupNodeList(rpc ndr.Invoker, hGroup mscmrp.HGROUP_RPC, multiSzNodeList []uint16, cchListSize ndr.DWORD) (Rpc_status ndr.DWORD, err error) {
	req := &apiSetGroupNodeListRequest{
		HGroup:          hGroup,
		MultiSzNodeList: multiSzNodeList,
		CchListSize:     cchListSize,
	}
	var resp apiSetGroupNodeListResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiSetGroupNodeList: %w", err)
		return
	}
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiSetGroupNodeList failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
