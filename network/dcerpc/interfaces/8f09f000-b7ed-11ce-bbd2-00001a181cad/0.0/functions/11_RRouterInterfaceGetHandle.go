package functions

// IDL source: [MS-RRASM] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rrasm/8e6e89fb-9c80-4a9a-a222-d7d8948244bb
// A fetched copy is kept at ms-rrasm.idl in the interface directory.

import (
	"fmt"

	dimsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/8f09f000-b7ed-11ce-bbd2-00001a181cad/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rRouterInterfaceGetHandleRequest carries the [in] parameters of RRouterInterfaceGetHandle.
type rRouterInterfaceGetHandleRequest struct {
	LpwsInterfaceName        ndr.WSTR
	PhInterface              ndr.DWORD
	FIncludeClientInterfaces ndr.DWORD
}

func (*rRouterInterfaceGetHandleRequest) Opnum() uint16 { return dimsvc.OpnumRRouterInterfaceGetHandle }

// rRouterInterfaceGetHandleResponse carries the [out] parameters and return value of RRouterInterfaceGetHandle.
type rRouterInterfaceGetHandleResponse struct {
	PhInterface ndr.DWORD
	Status      ndr.DWORD `ndr:"retval"`
}

// RRouterInterfaceGetHandle calls RRouterInterfaceGetHandle (opnum 11) ([MS-RRASM] — verify the parameter
// modeling and status handling).
func RRouterInterfaceGetHandle(rpc ndr.Invoker, lpwsInterfaceName ndr.WSTR, phInterface ndr.DWORD, fIncludeClientInterfaces ndr.DWORD) (PhInterface ndr.DWORD, err error) {
	req := &rRouterInterfaceGetHandleRequest{
		LpwsInterfaceName:        lpwsInterfaceName,
		PhInterface:              phInterface,
		FIncludeClientInterfaces: fIncludeClientInterfaces,
	}
	var resp rRouterInterfaceGetHandleResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRouterInterfaceGetHandle: %w", err)
		return
	}
	PhInterface = resp.PhInterface
	if uint32(resp.Status) != dimsvc.StatusSuccess {
		err = fmt.Errorf("RRouterInterfaceGetHandle failed: %s", dimsvc.StatusString(uint32(resp.Status)))
	}
	return
}
