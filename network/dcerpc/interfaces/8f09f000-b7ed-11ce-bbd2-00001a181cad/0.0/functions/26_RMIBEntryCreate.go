package functions

import (
	"fmt"

	dimsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/8f09f000-b7ed-11ce-bbd2-00001a181cad/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrrasm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rrasm"
)

// rMIBEntryCreateRequest carries the [in] parameters of RMIBEntryCreate.
type rMIBEntryCreateRequest struct {
	DwPid        ndr.DWORD
	DwRoutingPid ndr.DWORD
	PInfoStuct   msrrasm.DIM_MIB_ENTRY_CONTAINER
}

func (*rMIBEntryCreateRequest) Opnum() uint16 { return dimsvc.OpnumRMIBEntryCreate }

// rMIBEntryCreateResponse carries the [out] parameters and return value of RMIBEntryCreate.
type rMIBEntryCreateResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RMIBEntryCreate calls RMIBEntryCreate (opnum 26) ([MS-RRASM] — verify the parameter
// modeling and status handling).
func RMIBEntryCreate(rpc ndr.Invoker, dwPid ndr.DWORD, dwRoutingPid ndr.DWORD, pInfoStuct msrrasm.DIM_MIB_ENTRY_CONTAINER) (err error) {
	req := &rMIBEntryCreateRequest{
		DwPid:        dwPid,
		DwRoutingPid: dwRoutingPid,
		PInfoStuct:   pInfoStuct,
	}
	var resp rMIBEntryCreateResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RMIBEntryCreate: %w", err)
		return
	}
	if uint32(resp.Status) != dimsvc.StatusSuccess {
		err = fmt.Errorf("RMIBEntryCreate failed: %s", dimsvc.StatusString(uint32(resp.Status)))
	}
	return
}
