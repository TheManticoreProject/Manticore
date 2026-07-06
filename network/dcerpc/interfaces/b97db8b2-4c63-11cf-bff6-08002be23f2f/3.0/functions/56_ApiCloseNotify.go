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

// apiCloseNotifyRequest carries the [in] parameters of ApiCloseNotify.
type apiCloseNotifyRequest struct {
	Notify mscmrp.HNOTIFY_RPC
}

func (*apiCloseNotifyRequest) Opnum() uint16 { return clusapi.OpnumApiCloseNotify }

// apiCloseNotifyResponse carries the [out] parameters and return value of ApiCloseNotify.
type apiCloseNotifyResponse struct {
	Notify mscmrp.HNOTIFY_RPC
	Status ndr.DWORD `ndr:"retval"`
}

// ApiCloseNotify calls ApiCloseNotify (opnum 56) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiCloseNotify(rpc ndr.Invoker, notify mscmrp.HNOTIFY_RPC) (Notify mscmrp.HNOTIFY_RPC, err error) {
	req := &apiCloseNotifyRequest{
		Notify: notify,
	}
	var resp apiCloseNotifyResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiCloseNotify: %w", err)
		return
	}
	Notify = resp.Notify
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiCloseNotify failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
