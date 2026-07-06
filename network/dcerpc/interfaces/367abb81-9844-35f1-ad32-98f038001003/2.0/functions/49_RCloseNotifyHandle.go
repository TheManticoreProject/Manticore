package functions

// IDL source: [MS-SCMR] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-scmr/19168537-40b5-4d7a-99e0-d77f0f5e0241
// A fetched copy is kept at ms-scmr.idl in the interface directory.

import (
	"fmt"

	svcctl "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/367abb81-9844-35f1-ad32-98f038001003/2.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msscmr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-scmr"
)

// rCloseNotifyHandleRequest carries the [in] parameters of RCloseNotifyHandle.
type rCloseNotifyHandleRequest struct {
	PhNotify msscmr.LPSC_NOTIFY_RPC_HANDLE
}

func (*rCloseNotifyHandleRequest) Opnum() uint16 { return svcctl.OpnumRCloseNotifyHandle }

// rCloseNotifyHandleResponse carries the [out] parameters and return value of RCloseNotifyHandle.
type rCloseNotifyHandleResponse struct {
	PhNotify   msscmr.LPSC_NOTIFY_RPC_HANDLE
	PfApcFired ndr.BOOL
	Status     ndr.DWORD `ndr:"retval"`
}

// RCloseNotifyHandle calls RCloseNotifyHandle (opnum 49) ([MS-SCMR] — verify the parameter
// modeling and status handling).
func RCloseNotifyHandle(rpc ndr.Invoker, phNotify msscmr.LPSC_NOTIFY_RPC_HANDLE) (PhNotify msscmr.LPSC_NOTIFY_RPC_HANDLE, PfApcFired ndr.BOOL, err error) {
	req := &rCloseNotifyHandleRequest{
		PhNotify: phNotify,
	}
	var resp rCloseNotifyHandleResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RCloseNotifyHandle: %w", err)
		return
	}
	PhNotify = resp.PhNotify
	PfApcFired = resp.PfApcFired
	if uint32(resp.Status) != svcctl.StatusSuccess {
		err = fmt.Errorf("RCloseNotifyHandle failed: %s", svcctl.StatusString(uint32(resp.Status)))
	}
	return
}
