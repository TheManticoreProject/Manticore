package functions

// IDL source: [MS-RRASM] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rrasm/8e6e89fb-9c80-4a9a-a222-d7d8948244bb
// A fetched copy is kept at ms-rrasm.idl in the interface directory.

import (
	"fmt"

	dimsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/8f09f000-b7ed-11ce-bbd2-00001a181cad/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrrasm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rrasm"
)

// rRouterInterfaceGetCredentialsExRequest carries the [in] parameters of RRouterInterfaceGetCredentialsEx.
type rRouterInterfaceGetCredentialsExRequest struct {
	DwLevel     ndr.DWORD
	PInfoStruct msrrasm.DIM_INFORMATION_CONTAINER
	HInterface  ndr.DWORD
}

func (*rRouterInterfaceGetCredentialsExRequest) Opnum() uint16 {
	return dimsvc.OpnumRRouterInterfaceGetCredentialsEx
}

// rRouterInterfaceGetCredentialsExResponse carries the [out] parameters and return value of RRouterInterfaceGetCredentialsEx.
type rRouterInterfaceGetCredentialsExResponse struct {
	PInfoStruct msrrasm.DIM_INFORMATION_CONTAINER
	Status      ndr.DWORD `ndr:"retval"`
}

// RRouterInterfaceGetCredentialsEx calls RRouterInterfaceGetCredentialsEx (opnum 41) ([MS-RRASM] — verify the parameter
// modeling and status handling).
func RRouterInterfaceGetCredentialsEx(rpc ndr.Invoker, dwLevel ndr.DWORD, pInfoStruct msrrasm.DIM_INFORMATION_CONTAINER, hInterface ndr.DWORD) (PInfoStruct msrrasm.DIM_INFORMATION_CONTAINER, err error) {
	req := &rRouterInterfaceGetCredentialsExRequest{
		DwLevel:     dwLevel,
		PInfoStruct: pInfoStruct,
		HInterface:  hInterface,
	}
	var resp rRouterInterfaceGetCredentialsExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRouterInterfaceGetCredentialsEx: %w", err)
		return
	}
	PInfoStruct = resp.PInfoStruct
	if uint32(resp.Status) != dimsvc.StatusSuccess {
		err = fmt.Errorf("RRouterInterfaceGetCredentialsEx failed: %s", dimsvc.StatusString(uint32(resp.Status)))
	}
	return
}
