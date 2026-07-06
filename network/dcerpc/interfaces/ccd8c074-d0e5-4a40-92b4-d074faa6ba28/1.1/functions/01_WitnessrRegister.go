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

// witnessrRegisterRequest carries the [in] parameters of WitnessrRegister.
type witnessrRegisterRequest struct {
	Version            ndr.DWORD
	NetName            *ndr.WSTR `ndr:"unique"`
	IpAddress          *ndr.WSTR `ndr:"unique"`
	ClientComputerName *ndr.WSTR `ndr:"unique"`
}

func (*witnessrRegisterRequest) Opnum() uint16 { return Witness.OpnumWitnessrRegister }

// witnessrRegisterResponse carries the [out] parameters and return value of WitnessrRegister.
type witnessrRegisterResponse struct {
	PpContext msswn.PPCONTEXT_HANDLE
	Status    ndr.DWORD `ndr:"retval"`
}

// WitnessrRegister calls WitnessrRegister (opnum 1) ([MS-SWN] 3.1.4.2). It registers the client for resource state-change notifications on a NetName/IpAddress (Witness protocol v1).
func WitnessrRegister(rpc ndr.Invoker, version ndr.DWORD, netName *ndr.WSTR, ipAddress *ndr.WSTR, clientComputerName *ndr.WSTR) (PpContext msswn.PPCONTEXT_HANDLE, err error) {
	req := &witnessrRegisterRequest{
		Version:            version,
		NetName:            netName,
		IpAddress:          ipAddress,
		ClientComputerName: clientComputerName,
	}
	var resp witnessrRegisterResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("WitnessrRegister: %w", err)
		return
	}
	PpContext = resp.PpContext
	if uint32(resp.Status) != Witness.StatusSuccess {
		err = fmt.Errorf("WitnessrRegister failed: %s", Witness.StatusString(uint32(resp.Status)))
	}
	return
}
