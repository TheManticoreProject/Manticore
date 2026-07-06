package functions

// IDL source: [MS-SCMR] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-scmr/19168537-40b5-4d7a-99e0-d77f0f5e0241
// A fetched copy is kept at ms-scmr.idl in the interface directory.

import (
	"fmt"

	svcctl "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/367abb81-9844-35f1-ad32-98f038001003/2.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msscmr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-scmr"
)

// rStartServiceWRequest carries the [in] parameters of RStartServiceW.
type rStartServiceWRequest struct {
	HService msscmr.SC_RPC_HANDLE
	Argc     ndr.DWORD
	Argv     []msscmr.STRING_PTRSW `ndr:"unique,size_is=Argc"`
}

func (*rStartServiceWRequest) Opnum() uint16 { return svcctl.OpnumRStartServiceW }

// rStartServiceWResponse carries the [out] parameters and return value of RStartServiceW.
type rStartServiceWResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RStartServiceW calls RStartServiceW (opnum 19) ([MS-SCMR] — verify the parameter
// modeling and status handling).
func RStartServiceW(rpc ndr.Invoker, hService msscmr.SC_RPC_HANDLE, argc ndr.DWORD, argv []msscmr.STRING_PTRSW) (err error) {
	req := &rStartServiceWRequest{
		HService: hService,
		Argc:     argc,
		Argv:     argv,
	}
	var resp rStartServiceWResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RStartServiceW: %w", err)
		return
	}
	if uint32(resp.Status) != svcctl.StatusSuccess {
		err = fmt.Errorf("RStartServiceW failed: %s", svcctl.StatusString(uint32(resp.Status)))
	}
	return
}
