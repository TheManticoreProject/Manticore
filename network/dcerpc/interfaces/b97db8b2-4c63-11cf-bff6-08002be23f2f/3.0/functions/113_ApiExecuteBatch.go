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

// apiExecuteBatchRequest carries the [in] parameters of ApiExecuteBatch.
type apiExecuteBatchRequest struct {
	HKey   mscmrp.HKEY_RPC
	CbData ndr.DWORD
	LpData []uint8 `ndr:"ref,size_is=CbData"`
}

func (*apiExecuteBatchRequest) Opnum() uint16 { return clusapi.OpnumApiExecuteBatch }

// apiExecuteBatchResponse carries the [out] parameters and return value of ApiExecuteBatch.
type apiExecuteBatchResponse struct {
	PdwFailedCommand int32
	Rpc_status       ndr.DWORD
	Status           ndr.DWORD `ndr:"retval"`
}

// ApiExecuteBatch calls ApiExecuteBatch (opnum 113) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiExecuteBatch(rpc ndr.Invoker, hKey mscmrp.HKEY_RPC, cbData ndr.DWORD, lpData []uint8) (PdwFailedCommand int32, Rpc_status ndr.DWORD, err error) {
	req := &apiExecuteBatchRequest{
		HKey:   hKey,
		CbData: cbData,
		LpData: lpData,
	}
	var resp apiExecuteBatchResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiExecuteBatch: %w", err)
		return
	}
	PdwFailedCommand = resp.PdwFailedCommand
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiExecuteBatch failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
