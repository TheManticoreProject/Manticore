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

// witnessrUnRegisterExRequest carries the [in] parameters of WitnessrUnRegisterEx.
type witnessrUnRegisterExRequest struct {
	PpContext msswn.PPCONTEXT_HANDLE
}

func (*witnessrUnRegisterExRequest) Opnum() uint16 { return Witness.OpnumWitnessrUnRegisterEx }

// witnessrUnRegisterExResponse carries the [out] parameters and return value of WitnessrUnRegisterEx.
type witnessrUnRegisterExResponse struct {
	PpContext msswn.PPCONTEXT_HANDLE
	Status    ndr.DWORD `ndr:"retval"`
}

// WitnessrUnRegisterEx calls WitnessrUnRegisterEx (opnum 5) ([MS-SWN] 3.1.4.6). It removes a registration created by WitnessrRegisterEx (Witness protocol v2 only).
func WitnessrUnRegisterEx(rpc ndr.Invoker, ppContext msswn.PPCONTEXT_HANDLE) (PpContext msswn.PPCONTEXT_HANDLE, err error) {
	req := &witnessrUnRegisterExRequest{
		PpContext: ppContext,
	}
	var resp witnessrUnRegisterExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("WitnessrUnRegisterEx: %w", err)
		return
	}
	PpContext = resp.PpContext
	if uint32(resp.Status) != Witness.StatusSuccess {
		err = fmt.Errorf("WitnessrUnRegisterEx failed: %s", Witness.StatusString(uint32(resp.Status)))
	}
	return
}
