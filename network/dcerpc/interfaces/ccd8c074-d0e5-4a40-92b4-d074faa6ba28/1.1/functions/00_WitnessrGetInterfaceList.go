package functions

// IDL source: [MS-SWN] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-swn/ccebaef8-60b0-4847-9ed7-2519d2b6ef19
// A fetched copy is kept at ms-swn.idl in the interface directory.

import (
	"fmt"

	Witness "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/ccd8c074-d0e5-4a40-92b4-d074faa6ba28/1.1"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msswn "github.com/TheManticoreProject/Manticore/windows/protocols/ms-swn"
)

// witnessrGetInterfaceListRequest carries the [in] parameters of WitnessrGetInterfaceList.
type witnessrGetInterfaceListRequest struct {
}

func (*witnessrGetInterfaceListRequest) Opnum() uint16 { return Witness.OpnumWitnessrGetInterfaceList }

// witnessrGetInterfaceListResponse carries the [out] parameters and return value of WitnessrGetInterfaceList.
type witnessrGetInterfaceListResponse struct {
	InterfaceList *msswn.WITNESS_INTERFACE_LIST `ndr:"unique"`
	Status        ndr.DWORD                     `ndr:"retval"`
}

// WitnessrGetInterfaceList calls WitnessrGetInterfaceList (opnum 0) ([MS-SWN] 3.1.4.1). It returns the list of witness-capable interfaces the server exposes.
func WitnessrGetInterfaceList(rpc ndr.Invoker) (InterfaceList *msswn.WITNESS_INTERFACE_LIST, err error) {
	req := &witnessrGetInterfaceListRequest{}
	var resp witnessrGetInterfaceListResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("WitnessrGetInterfaceList: %w", err)
		return
	}
	InterfaceList = resp.InterfaceList
	if uint32(resp.Status) != Witness.StatusSuccess {
		err = fmt.Errorf("WitnessrGetInterfaceList failed: %s", Witness.StatusString(uint32(resp.Status)))
	}
	return
}
