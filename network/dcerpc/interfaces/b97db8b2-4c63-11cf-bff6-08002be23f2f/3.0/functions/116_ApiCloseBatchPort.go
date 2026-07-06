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

// apiCloseBatchPortRequest carries the [in] parameters of ApiCloseBatchPort.
type apiCloseBatchPortRequest struct {
	PhBatchPort mscmrp.HBATCH_PORT_RPC
}

func (*apiCloseBatchPortRequest) Opnum() uint16 { return clusapi.OpnumApiCloseBatchPort }

// apiCloseBatchPortResponse carries the [out] parameters and return value of ApiCloseBatchPort.
type apiCloseBatchPortResponse struct {
	PhBatchPort mscmrp.HBATCH_PORT_RPC
	Status      ndr.DWORD `ndr:"retval"`
}

// ApiCloseBatchPort calls ApiCloseBatchPort (opnum 116) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiCloseBatchPort(rpc ndr.Invoker, phBatchPort mscmrp.HBATCH_PORT_RPC) (PhBatchPort mscmrp.HBATCH_PORT_RPC, err error) {
	req := &apiCloseBatchPortRequest{
		PhBatchPort: phBatchPort,
	}
	var resp apiCloseBatchPortResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiCloseBatchPort: %w", err)
		return
	}
	PhBatchPort = resp.PhBatchPort
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiCloseBatchPort failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
