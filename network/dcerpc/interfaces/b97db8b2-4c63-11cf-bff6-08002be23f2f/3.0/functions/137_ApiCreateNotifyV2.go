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

// apiCreateNotifyV2Request carries the [in] parameters of ApiCreateNotifyV2.
type apiCreateNotifyV2Request struct {
}

func (*apiCreateNotifyV2Request) Opnum() uint16 { return clusapi.OpnumApiCreateNotifyV2 }

// apiCreateNotifyV2Response carries the [out] parameters and return value of ApiCreateNotifyV2.
type apiCreateNotifyV2Response struct {
	Rpc_error  ndr.DWORD
	Rpc_status ndr.DWORD
	Handle     mscmrp.HNOTIFY_RPC `ndr:"retval"`
}

// ApiCreateNotifyV2 calls ApiCreateNotifyV2 (opnum 137) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiCreateNotifyV2(rpc ndr.Invoker) (Handle mscmrp.HNOTIFY_RPC, Rpc_error ndr.DWORD, Rpc_status ndr.DWORD, err error) {
	req := &apiCreateNotifyV2Request{}
	var resp apiCreateNotifyV2Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiCreateNotifyV2: %w", err)
		return
	}
	Handle = resp.Handle
	Rpc_error = resp.Rpc_error
	Rpc_status = resp.Rpc_status
	if uint32(resp.Rpc_error) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiCreateNotifyV2 failed: %s", clusapi.StatusString(uint32(resp.Rpc_error)))
	}
	return
}
