package functions

import (
	"fmt"

	dimsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/8f09f000-b7ed-11ce-bbd2-00001a181cad/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrrasm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rrasm"
)

// rMIBEntryGetFirstRequest carries the [in] parameters of RMIBEntryGetFirst.
type rMIBEntryGetFirstRequest struct {
	DwPid        ndr.DWORD
	DwRoutingPid ndr.DWORD
	PInfoStuct   msrrasm.DIM_MIB_ENTRY_CONTAINER
}

func (*rMIBEntryGetFirstRequest) Opnum() uint16 { return dimsvc.OpnumRMIBEntryGetFirst }

// rMIBEntryGetFirstResponse carries the [out] parameters and return value of RMIBEntryGetFirst.
type rMIBEntryGetFirstResponse struct {
	PInfoStuct msrrasm.DIM_MIB_ENTRY_CONTAINER
	Status     ndr.DWORD `ndr:"retval"`
}

// RMIBEntryGetFirst calls RMIBEntryGetFirst (opnum 30) ([MS-RRASM] — verify the parameter
// modeling and status handling).
func RMIBEntryGetFirst(rpc ndr.Invoker, dwPid ndr.DWORD, dwRoutingPid ndr.DWORD, pInfoStuct msrrasm.DIM_MIB_ENTRY_CONTAINER) (PInfoStuct msrrasm.DIM_MIB_ENTRY_CONTAINER, err error) {
	req := &rMIBEntryGetFirstRequest{
		DwPid:        dwPid,
		DwRoutingPid: dwRoutingPid,
		PInfoStuct:   pInfoStuct,
	}
	var resp rMIBEntryGetFirstResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RMIBEntryGetFirst: %w", err)
		return
	}
	PInfoStuct = resp.PInfoStuct
	if uint32(resp.Status) != dimsvc.StatusSuccess {
		err = fmt.Errorf("RMIBEntryGetFirst failed: %s", dimsvc.StatusString(uint32(resp.Status)))
	}
	return
}
