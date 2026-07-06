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

// rRouterInterfaceTransportCreateRequest carries the [in] parameters of RRouterInterfaceTransportCreate.
type rRouterInterfaceTransportCreateRequest struct {
	DwTransportId     ndr.DWORD
	LpwsTransportName ndr.WSTR
	PInfoStruct       msrrasm.DIM_INTERFACE_CONTAINER
	LpwsDLLPath       ndr.WSTR
}

func (*rRouterInterfaceTransportCreateRequest) Opnum() uint16 {
	return dimsvc.OpnumRRouterInterfaceTransportCreate
}

// rRouterInterfaceTransportCreateResponse carries the [out] parameters and return value of RRouterInterfaceTransportCreate.
type rRouterInterfaceTransportCreateResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RRouterInterfaceTransportCreate calls RRouterInterfaceTransportCreate (opnum 37) ([MS-RRASM] — verify the parameter
// modeling and status handling).
func RRouterInterfaceTransportCreate(rpc ndr.Invoker, dwTransportId ndr.DWORD, lpwsTransportName ndr.WSTR, pInfoStruct msrrasm.DIM_INTERFACE_CONTAINER, lpwsDLLPath ndr.WSTR) (err error) {
	req := &rRouterInterfaceTransportCreateRequest{
		DwTransportId:     dwTransportId,
		LpwsTransportName: lpwsTransportName,
		PInfoStruct:       pInfoStruct,
		LpwsDLLPath:       lpwsDLLPath,
	}
	var resp rRouterInterfaceTransportCreateResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRouterInterfaceTransportCreate: %w", err)
		return
	}
	if uint32(resp.Status) != dimsvc.StatusSuccess {
		err = fmt.Errorf("RRouterInterfaceTransportCreate failed: %s", dimsvc.StatusString(uint32(resp.Status)))
	}
	return
}
