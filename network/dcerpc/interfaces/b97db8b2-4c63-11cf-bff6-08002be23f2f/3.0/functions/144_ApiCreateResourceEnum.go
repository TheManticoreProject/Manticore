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

// apiCreateResourceEnumRequest carries the [in] parameters of ApiCreateResourceEnum.
type apiCreateResourceEnumRequest struct {
	HCluster       mscmrp.HCLUSTER_RPC
	PProperties    []uint8 `ndr:"ref,size_is=CbProperties"`
	CbProperties   ndr.DWORD
	PRoProperties  []uint8 `ndr:"ref,size_is=CbRoProperties"`
	CbRoProperties ndr.DWORD
}

func (*apiCreateResourceEnumRequest) Opnum() uint16 { return clusapi.OpnumApiCreateResourceEnum }

// apiCreateResourceEnumResponse carries the [out] parameters and return value of ApiCreateResourceEnum.
type apiCreateResourceEnumResponse struct {
	PpResultList *mscmrp.RESOURCE_ENUM_LIST `ndr:"unique"`
	Rpc_status   ndr.DWORD
	Status       ndr.DWORD `ndr:"retval"`
}

// ApiCreateResourceEnum calls ApiCreateResourceEnum (opnum 144) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiCreateResourceEnum(rpc ndr.Invoker, hCluster mscmrp.HCLUSTER_RPC, pProperties []uint8, cbProperties ndr.DWORD, pRoProperties []uint8, cbRoProperties ndr.DWORD) (PpResultList *mscmrp.RESOURCE_ENUM_LIST, Rpc_status ndr.DWORD, err error) {
	req := &apiCreateResourceEnumRequest{
		HCluster:       hCluster,
		PProperties:    pProperties,
		CbProperties:   cbProperties,
		PRoProperties:  pRoProperties,
		CbRoProperties: cbRoProperties,
	}
	var resp apiCreateResourceEnumResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiCreateResourceEnum: %w", err)
		return
	}
	PpResultList = resp.PpResultList
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiCreateResourceEnum failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
