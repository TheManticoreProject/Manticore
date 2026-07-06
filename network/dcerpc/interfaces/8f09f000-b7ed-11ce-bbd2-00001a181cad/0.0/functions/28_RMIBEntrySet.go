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

// rMIBEntrySetRequest carries the [in] parameters of RMIBEntrySet.
type rMIBEntrySetRequest struct {
	DwPid        ndr.DWORD
	DwRoutingPid ndr.DWORD
	PInfoStuct   msrrasm.DIM_MIB_ENTRY_CONTAINER
}

func (*rMIBEntrySetRequest) Opnum() uint16 { return dimsvc.OpnumRMIBEntrySet }

// rMIBEntrySetResponse carries the [out] parameters and return value of RMIBEntrySet.
type rMIBEntrySetResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RMIBEntrySet calls RMIBEntrySet (opnum 28) ([MS-RRASM] — verify the parameter
// modeling and status handling).
func RMIBEntrySet(rpc ndr.Invoker, dwPid ndr.DWORD, dwRoutingPid ndr.DWORD, pInfoStuct msrrasm.DIM_MIB_ENTRY_CONTAINER) (err error) {
	req := &rMIBEntrySetRequest{
		DwPid:        dwPid,
		DwRoutingPid: dwRoutingPid,
		PInfoStuct:   pInfoStuct,
	}
	var resp rMIBEntrySetResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RMIBEntrySet: %w", err)
		return
	}
	if uint32(resp.Status) != dimsvc.StatusSuccess {
		err = fmt.Errorf("RMIBEntrySet failed: %s", dimsvc.StatusString(uint32(resp.Status)))
	}
	return
}
