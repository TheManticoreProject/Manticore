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

// rRouterInterfaceCreateRequest carries the [in] parameters of RRouterInterfaceCreate.
type rRouterInterfaceCreateRequest struct {
	DwLevel     ndr.DWORD
	PInfoStruct msrrasm.DIM_INFORMATION_CONTAINER
	PhInterface ndr.DWORD
}

func (*rRouterInterfaceCreateRequest) Opnum() uint16 { return dimsvc.OpnumRRouterInterfaceCreate }

// rRouterInterfaceCreateResponse carries the [out] parameters and return value of RRouterInterfaceCreate.
type rRouterInterfaceCreateResponse struct {
	PhInterface ndr.DWORD
	Status      ndr.DWORD `ndr:"retval"`
}

// RRouterInterfaceCreate calls RRouterInterfaceCreate (opnum 12) ([MS-RRASM] — verify the parameter
// modeling and status handling).
func RRouterInterfaceCreate(rpc ndr.Invoker, dwLevel ndr.DWORD, pInfoStruct msrrasm.DIM_INFORMATION_CONTAINER, phInterface ndr.DWORD) (PhInterface ndr.DWORD, err error) {
	req := &rRouterInterfaceCreateRequest{
		DwLevel:     dwLevel,
		PInfoStruct: pInfoStruct,
		PhInterface: phInterface,
	}
	var resp rRouterInterfaceCreateResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRouterInterfaceCreate: %w", err)
		return
	}
	PhInterface = resp.PhInterface
	if uint32(resp.Status) != dimsvc.StatusSuccess {
		err = fmt.Errorf("RRouterInterfaceCreate failed: %s", dimsvc.StatusString(uint32(resp.Status)))
	}
	return
}
