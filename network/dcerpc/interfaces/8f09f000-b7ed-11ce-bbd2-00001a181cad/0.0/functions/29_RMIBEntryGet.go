package functions

import (
	"fmt"

	dimsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/8f09f000-b7ed-11ce-bbd2-00001a181cad/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrrasm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rrasm"
)

// rMIBEntryGetRequest carries the [in] parameters of RMIBEntryGet.
type rMIBEntryGetRequest struct {
	DwPid        ndr.DWORD
	DwRoutingPid ndr.DWORD
	PInfoStuct   msrrasm.DIM_MIB_ENTRY_CONTAINER
}

func (*rMIBEntryGetRequest) Opnum() uint16 { return dimsvc.OpnumRMIBEntryGet }

// rMIBEntryGetResponse carries the [out] parameters and return value of RMIBEntryGet.
type rMIBEntryGetResponse struct {
	PInfoStuct msrrasm.DIM_MIB_ENTRY_CONTAINER
	Status     ndr.DWORD `ndr:"retval"`
}

// RMIBEntryGet calls RMIBEntryGet (opnum 29) ([MS-RRASM] — verify the parameter
// modeling and status handling).
func RMIBEntryGet(rpc ndr.Invoker, dwPid ndr.DWORD, dwRoutingPid ndr.DWORD, pInfoStuct msrrasm.DIM_MIB_ENTRY_CONTAINER) (PInfoStuct msrrasm.DIM_MIB_ENTRY_CONTAINER, err error) {
	req := &rMIBEntryGetRequest{
		DwPid:        dwPid,
		DwRoutingPid: dwRoutingPid,
		PInfoStuct:   pInfoStuct,
	}
	var resp rMIBEntryGetResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RMIBEntryGet: %w", err)
		return
	}
	PInfoStuct = resp.PInfoStuct
	if uint32(resp.Status) != dimsvc.StatusSuccess {
		err = fmt.Errorf("RMIBEntryGet failed: %s", dimsvc.StatusString(uint32(resp.Status)))
	}
	return
}
