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

// apiCreateBatchPortRequest carries the [in] parameters of ApiCreateBatchPort.
type apiCreateBatchPortRequest struct {
	HKey mscmrp.HKEY_RPC
}

func (*apiCreateBatchPortRequest) Opnum() uint16 { return clusapi.OpnumApiCreateBatchPort }

// apiCreateBatchPortResponse carries the [out] parameters and return value of ApiCreateBatchPort.
type apiCreateBatchPortResponse struct {
	PhBatchPort mscmrp.HBATCH_PORT_RPC
	Rpc_status  ndr.DWORD
	Status      ndr.DWORD `ndr:"retval"`
}

// ApiCreateBatchPort calls ApiCreateBatchPort (opnum 114) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiCreateBatchPort(rpc ndr.Invoker, hKey mscmrp.HKEY_RPC) (PhBatchPort mscmrp.HBATCH_PORT_RPC, Rpc_status ndr.DWORD, err error) {
	req := &apiCreateBatchPortRequest{
		HKey: hKey,
	}
	var resp apiCreateBatchPortResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiCreateBatchPort: %w", err)
		return
	}
	PhBatchPort = resp.PhBatchPort
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiCreateBatchPort failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
