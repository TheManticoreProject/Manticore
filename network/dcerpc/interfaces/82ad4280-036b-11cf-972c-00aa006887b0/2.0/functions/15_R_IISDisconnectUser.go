package functions

// IDL source: [MS-IRP] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-irp/ed7e5940-9700-4a1f-8555-de29f99fe115
// A fetched copy is kept at ms-irp.idl in the interface directory.

import (
	"fmt"

	inetinfo "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/82ad4280-036b-11cf-972c-00aa006887b0/2.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// r_IISDisconnectUserRequest carries the [in] parameters of R_IISDisconnectUser.
type r_IISDisconnectUserRequest struct {
	PszServer   *ndr.WSTR `ndr:"unique"`
	DwServiceId ndr.DWORD
	DwInstance  ndr.DWORD
	DwIdUser    ndr.DWORD
}

func (*r_IISDisconnectUserRequest) Opnum() uint16 { return inetinfo.OpnumR_IISDisconnectUser }

// r_IISDisconnectUserResponse carries the [out] parameters and return value of R_IISDisconnectUser.
type r_IISDisconnectUserResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_IISDisconnectUser calls R_IISDisconnectUser (opnum 15) ([MS-IRP] — verify the parameter
// modeling and status handling).
func R_IISDisconnectUser(rpc ndr.Invoker, pszServer *ndr.WSTR, dwServiceId ndr.DWORD, dwInstance ndr.DWORD, dwIdUser ndr.DWORD) (err error) {
	req := &r_IISDisconnectUserRequest{
		PszServer:   pszServer,
		DwServiceId: dwServiceId,
		DwInstance:  dwInstance,
		DwIdUser:    dwIdUser,
	}
	var resp r_IISDisconnectUserResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_IISDisconnectUser: %w", err)
		return
	}
	if uint32(resp.Status) != inetinfo.ErrorSuccess {
		err = fmt.Errorf("R_IISDisconnectUser failed: %s", inetinfo.StatusString(uint32(resp.Status)))
	}
	return
}
