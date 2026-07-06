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

// witnessrRegisterExRequest carries the [in] parameters of WitnessrRegisterEx.
type witnessrRegisterExRequest struct {
	Version            ndr.DWORD
	NetName            *ndr.WSTR `ndr:"unique"`
	ShareName          *ndr.WSTR `ndr:"unique"`
	IpAddress          *ndr.WSTR `ndr:"unique"`
	ClientComputerName *ndr.WSTR `ndr:"unique"`
	Flags              ndr.DWORD
	KeepAliveTimeout   ndr.DWORD
}

func (*witnessrRegisterExRequest) Opnum() uint16 { return Witness.OpnumWitnessrRegisterEx }

// witnessrRegisterExResponse carries the [out] parameters and return value of WitnessrRegisterEx.
type witnessrRegisterExResponse struct {
	PpContext msswn.PPCONTEXT_HANDLE
	Status    ndr.DWORD `ndr:"retval"`
}

// WitnessrRegisterEx calls WitnessrRegisterEx (opnum 4) ([MS-SWN] 3.1.4.5). It registers the client for notifications on a specific share with optional flags and keep-alive (Witness protocol v2 only).
func WitnessrRegisterEx(rpc ndr.Invoker, version ndr.DWORD, netName *ndr.WSTR, shareName *ndr.WSTR, ipAddress *ndr.WSTR, clientComputerName *ndr.WSTR, flags ndr.DWORD, keepAliveTimeout ndr.DWORD) (PpContext msswn.PPCONTEXT_HANDLE, err error) {
	req := &witnessrRegisterExRequest{
		Version:            version,
		NetName:            netName,
		ShareName:          shareName,
		IpAddress:          ipAddress,
		ClientComputerName: clientComputerName,
		Flags:              flags,
		KeepAliveTimeout:   keepAliveTimeout,
	}
	var resp witnessrRegisterExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("WitnessrRegisterEx: %w", err)
		return
	}
	PpContext = resp.PpContext
	if uint32(resp.Status) != Witness.StatusSuccess {
		err = fmt.Errorf("WitnessrRegisterEx failed: %s", Witness.StatusString(uint32(resp.Status)))
	}
	return
}
