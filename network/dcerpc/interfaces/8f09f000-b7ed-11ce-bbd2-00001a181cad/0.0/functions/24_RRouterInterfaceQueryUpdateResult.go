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

// rRouterInterfaceQueryUpdateResultRequest carries the [in] parameters of RRouterInterfaceQueryUpdateResult.
type rRouterInterfaceQueryUpdateResultRequest struct {
	HInterface    ndr.DWORD
	DwTransportId ndr.DWORD
}

func (*rRouterInterfaceQueryUpdateResultRequest) Opnum() uint16 {
	return dimsvc.OpnumRRouterInterfaceQueryUpdateResult
}

// rRouterInterfaceQueryUpdateResultResponse carries the [out] parameters and return value of RRouterInterfaceQueryUpdateResult.
type rRouterInterfaceQueryUpdateResultResponse struct {
	PUpdateResult ndr.DWORD
	Status        ndr.DWORD `ndr:"retval"`
}

// RRouterInterfaceQueryUpdateResult calls RRouterInterfaceQueryUpdateResult (opnum 24) ([MS-RRASM] — verify the parameter
// modeling and status handling).
func RRouterInterfaceQueryUpdateResult(rpc ndr.Invoker, hInterface ndr.DWORD, dwTransportId ndr.DWORD) (PUpdateResult ndr.DWORD, err error) {
	req := &rRouterInterfaceQueryUpdateResultRequest{
		HInterface:    hInterface,
		DwTransportId: dwTransportId,
	}
	var resp rRouterInterfaceQueryUpdateResultResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRouterInterfaceQueryUpdateResult: %w", err)
		return
	}
	PUpdateResult = resp.PUpdateResult
	if uint32(resp.Status) != dimsvc.StatusSuccess {
		err = fmt.Errorf("RRouterInterfaceQueryUpdateResult failed: %s", dimsvc.StatusString(uint32(resp.Status)))
	}
	return
}
