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

// apiCreateNotifyRequest carries the [in] parameters of ApiCreateNotify.
type apiCreateNotifyRequest struct {
}

func (*apiCreateNotifyRequest) Opnum() uint16 { return clusapi.OpnumApiCreateNotify }

// apiCreateNotifyResponse carries the [out] parameters and return value of ApiCreateNotify.
type apiCreateNotifyResponse struct {
	Status     ndr.DWORD
	Rpc_status ndr.DWORD
	Handle     mscmrp.HNOTIFY_RPC `ndr:"retval"`
}

// ApiCreateNotify calls ApiCreateNotify (opnum 55) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiCreateNotify(rpc ndr.Invoker) (Handle mscmrp.HNOTIFY_RPC, Status ndr.DWORD, Rpc_status ndr.DWORD, err error) {
	req := &apiCreateNotifyRequest{}
	var resp apiCreateNotifyResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiCreateNotify: %w", err)
		return
	}
	Handle = resp.Handle
	Status = resp.Status
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiCreateNotify failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
