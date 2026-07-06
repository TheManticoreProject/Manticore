package functions

import (
	"fmt"

	dimsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/8f09f000-b7ed-11ce-bbd2-00001a181cad/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrrasm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rrasm"
)

// rMIBEntryDeleteRequest carries the [in] parameters of RMIBEntryDelete.
type rMIBEntryDeleteRequest struct {
	DwPid        ndr.DWORD
	DwRoutingPid ndr.DWORD
	PInfoStuct   msrrasm.DIM_MIB_ENTRY_CONTAINER
}

func (*rMIBEntryDeleteRequest) Opnum() uint16 { return dimsvc.OpnumRMIBEntryDelete }

// rMIBEntryDeleteResponse carries the [out] parameters and return value of RMIBEntryDelete.
type rMIBEntryDeleteResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RMIBEntryDelete calls RMIBEntryDelete (opnum 27) ([MS-RRASM] — verify the parameter
// modeling and status handling).
func RMIBEntryDelete(rpc ndr.Invoker, dwPid ndr.DWORD, dwRoutingPid ndr.DWORD, pInfoStuct msrrasm.DIM_MIB_ENTRY_CONTAINER) (err error) {
	req := &rMIBEntryDeleteRequest{
		DwPid:        dwPid,
		DwRoutingPid: dwRoutingPid,
		PInfoStuct:   pInfoStuct,
	}
	var resp rMIBEntryDeleteResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RMIBEntryDelete: %w", err)
		return
	}
	if uint32(resp.Status) != dimsvc.StatusSuccess {
		err = fmt.Errorf("RMIBEntryDelete failed: %s", dimsvc.StatusString(uint32(resp.Status)))
	}
	return
}
