package functions

import (
	"fmt"

	dimsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/8f09f000-b7ed-11ce-bbd2-00001a181cad/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrrasm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rrasm"
)

// rMIBSetTrapInfoRequest carries the [in] parameters of RMIBSetTrapInfo.
type rMIBSetTrapInfoRequest struct {
	DwPid             ndr.DWORD
	DwRoutingPid      ndr.DWORD
	HEvent            ndr.DWORD
	DwClientProcessId ndr.DWORD
	PInfoStruct       msrrasm.DIM_MIB_ENTRY_CONTAINER
}

func (*rMIBSetTrapInfoRequest) Opnum() uint16 { return dimsvc.OpnumRMIBSetTrapInfo }

// rMIBSetTrapInfoResponse carries the [out] parameters and return value of RMIBSetTrapInfo.
type rMIBSetTrapInfoResponse struct {
	PInfoStruct msrrasm.DIM_MIB_ENTRY_CONTAINER
	Status      ndr.DWORD `ndr:"retval"`
}

// RMIBSetTrapInfo calls RMIBSetTrapInfo (opnum 33) ([MS-RRASM] — verify the parameter
// modeling and status handling).
func RMIBSetTrapInfo(rpc ndr.Invoker, dwPid ndr.DWORD, dwRoutingPid ndr.DWORD, hEvent ndr.DWORD, dwClientProcessId ndr.DWORD, pInfoStruct msrrasm.DIM_MIB_ENTRY_CONTAINER) (PInfoStruct msrrasm.DIM_MIB_ENTRY_CONTAINER, err error) {
	req := &rMIBSetTrapInfoRequest{
		DwPid:             dwPid,
		DwRoutingPid:      dwRoutingPid,
		HEvent:            hEvent,
		DwClientProcessId: dwClientProcessId,
		PInfoStruct:       pInfoStruct,
	}
	var resp rMIBSetTrapInfoResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RMIBSetTrapInfo: %w", err)
		return
	}
	PInfoStruct = resp.PInfoStruct
	if uint32(resp.Status) != dimsvc.StatusSuccess {
		err = fmt.Errorf("RMIBSetTrapInfo failed: %s", dimsvc.StatusString(uint32(resp.Status)))
	}
	return
}
