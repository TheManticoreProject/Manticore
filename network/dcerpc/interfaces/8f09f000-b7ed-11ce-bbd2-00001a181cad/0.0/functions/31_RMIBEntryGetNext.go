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

// rMIBEntryGetNextRequest carries the [in] parameters of RMIBEntryGetNext.
type rMIBEntryGetNextRequest struct {
	DwPid        ndr.DWORD
	DwRoutingPid ndr.DWORD
	PInfoStuct   msrrasm.DIM_MIB_ENTRY_CONTAINER
}

func (*rMIBEntryGetNextRequest) Opnum() uint16 { return dimsvc.OpnumRMIBEntryGetNext }

// rMIBEntryGetNextResponse carries the [out] parameters and return value of RMIBEntryGetNext.
type rMIBEntryGetNextResponse struct {
	PInfoStuct msrrasm.DIM_MIB_ENTRY_CONTAINER
	Status     ndr.DWORD `ndr:"retval"`
}

// RMIBEntryGetNext calls RMIBEntryGetNext (opnum 31) ([MS-RRASM] — verify the parameter
// modeling and status handling).
func RMIBEntryGetNext(rpc ndr.Invoker, dwPid ndr.DWORD, dwRoutingPid ndr.DWORD, pInfoStuct msrrasm.DIM_MIB_ENTRY_CONTAINER) (PInfoStuct msrrasm.DIM_MIB_ENTRY_CONTAINER, err error) {
	req := &rMIBEntryGetNextRequest{
		DwPid:        dwPid,
		DwRoutingPid: dwRoutingPid,
		PInfoStuct:   pInfoStuct,
	}
	var resp rMIBEntryGetNextResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RMIBEntryGetNext: %w", err)
		return
	}
	PInfoStuct = resp.PInfoStuct
	if uint32(resp.Status) != dimsvc.StatusSuccess {
		err = fmt.Errorf("RMIBEntryGetNext failed: %s", dimsvc.StatusString(uint32(resp.Status)))
	}
	return
}
