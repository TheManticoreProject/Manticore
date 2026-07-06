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

// rMIBGetTrapInfoRequest carries the [in] parameters of RMIBGetTrapInfo.
type rMIBGetTrapInfoRequest struct {
	DwPid        ndr.DWORD
	DwRoutingPid ndr.DWORD
	PInfoStruct  msrrasm.DIM_MIB_ENTRY_CONTAINER
}

func (*rMIBGetTrapInfoRequest) Opnum() uint16 { return dimsvc.OpnumRMIBGetTrapInfo }

// rMIBGetTrapInfoResponse carries the [out] parameters and return value of RMIBGetTrapInfo.
type rMIBGetTrapInfoResponse struct {
	PInfoStruct msrrasm.DIM_MIB_ENTRY_CONTAINER
	Status      ndr.DWORD `ndr:"retval"`
}

// RMIBGetTrapInfo calls RMIBGetTrapInfo (opnum 32) ([MS-RRASM] — verify the parameter
// modeling and status handling).
func RMIBGetTrapInfo(rpc ndr.Invoker, dwPid ndr.DWORD, dwRoutingPid ndr.DWORD, pInfoStruct msrrasm.DIM_MIB_ENTRY_CONTAINER) (PInfoStruct msrrasm.DIM_MIB_ENTRY_CONTAINER, err error) {
	req := &rMIBGetTrapInfoRequest{
		DwPid:        dwPid,
		DwRoutingPid: dwRoutingPid,
		PInfoStruct:  pInfoStruct,
	}
	var resp rMIBGetTrapInfoResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RMIBGetTrapInfo: %w", err)
		return
	}
	PInfoStruct = resp.PInfoStruct
	if uint32(resp.Status) != dimsvc.StatusSuccess {
		err = fmt.Errorf("RMIBGetTrapInfo failed: %s", dimsvc.StatusString(uint32(resp.Status)))
	}
	return
}
