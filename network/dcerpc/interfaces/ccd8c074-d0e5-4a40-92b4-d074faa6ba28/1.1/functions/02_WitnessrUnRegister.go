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

// witnessrUnRegisterRequest carries the [in] parameters of WitnessrUnRegister.
type witnessrUnRegisterRequest struct {
	PContext msswn.PCONTEXT_HANDLE
}

func (*witnessrUnRegisterRequest) Opnum() uint16 { return Witness.OpnumWitnessrUnRegister }

// witnessrUnRegisterResponse carries the [out] parameters and return value of WitnessrUnRegister.
type witnessrUnRegisterResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// WitnessrUnRegister calls WitnessrUnRegister (opnum 2) ([MS-SWN] 3.1.4.3). It removes a registration created by WitnessrRegister.
func WitnessrUnRegister(rpc ndr.Invoker, pContext msswn.PCONTEXT_HANDLE) (err error) {
	req := &witnessrUnRegisterRequest{
		PContext: pContext,
	}
	var resp witnessrUnRegisterResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("WitnessrUnRegister: %w", err)
		return
	}
	if uint32(resp.Status) != Witness.StatusSuccess {
		err = fmt.Errorf("WitnessrUnRegister failed: %s", Witness.StatusString(uint32(resp.Status)))
	}
	return
}
