package functions

// IDL source: [MS-RRASM] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rrasm/8e6e89fb-9c80-4a9a-a222-d7d8948244bb
// A fetched copy is kept at ms-rrasm.idl in the interface directory.

import (
	"fmt"

	dimsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/8f09f000-b7ed-11ce-bbd2-00001a181cad/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/windows/errors/win32"
)

// rRouterInterfaceDeleteRequest carries the [in] parameters of RRouterInterfaceDelete.
type rRouterInterfaceDeleteRequest struct {
	HInterface ndr.DWORD
}

func (*rRouterInterfaceDeleteRequest) Opnum() uint16 { return dimsvc.OpnumRRouterInterfaceDelete }

// rRouterInterfaceDeleteResponse carries the [out] parameters and return value of RRouterInterfaceDelete.
type rRouterInterfaceDeleteResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RRouterInterfaceDelete calls RRouterInterfaceDelete (opnum 15) ([MS-RRASM] — verify the parameter
// modeling and status handling).
func RRouterInterfaceDelete(rpc ndr.Invoker, hInterface ndr.DWORD) (err error) {
	req := &rRouterInterfaceDeleteRequest{
		HInterface: hInterface,
	}
	var resp rRouterInterfaceDeleteResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRouterInterfaceDelete: %w", err)
		return
	}
	if win32.WIN32_ERROR(resp.Status) != win32.ERROR_SUCCESS {
		err = fmt.Errorf("RRouterInterfaceDelete failed: %s", win32.WIN32_ERROR(resp.Status).String())
	}
	return
}
