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

// rCloseServiceHandleRequest carries the [in] parameters of RCloseServiceHandle.
type rCloseServiceHandleRequest struct {
	HSCObject msscmr.LPSC_RPC_HANDLE
}

func (*rCloseServiceHandleRequest) Opnum() uint16 { return svcctl.OpnumRCloseServiceHandle }

// rCloseServiceHandleResponse carries the [out] parameters and return value of RCloseServiceHandle.
type rCloseServiceHandleResponse struct {
	HSCObject msscmr.LPSC_RPC_HANDLE
	Status    ndr.DWORD `ndr:"retval"`
}

// RCloseServiceHandle calls RCloseServiceHandle (opnum 0) ([MS-SCMR] — verify the parameter
// modeling and status handling).
func RCloseServiceHandle(rpc ndr.Invoker, hSCObject msscmr.LPSC_RPC_HANDLE) (HSCObject msscmr.LPSC_RPC_HANDLE, err error) {
	req := &rCloseServiceHandleRequest{
		HSCObject: hSCObject,
	}
	var resp rCloseServiceHandleResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RCloseServiceHandle: %w", err)
		return
	}
	HSCObject = resp.HSCObject
	if uint32(resp.Status) != svcctl.StatusSuccess {
		err = fmt.Errorf("RCloseServiceHandle failed: %s", svcctl.StatusString(uint32(resp.Status)))
	}
	return
}
