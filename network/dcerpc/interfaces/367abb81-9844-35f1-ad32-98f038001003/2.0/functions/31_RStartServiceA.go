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

// rStartServiceARequest carries the [in] parameters of RStartServiceA.
type rStartServiceARequest struct {
	HService msscmr.SC_RPC_HANDLE
	Argc     ndr.DWORD
	Argv     []msscmr.STRING_PTRSA `ndr:"unique,size_is=Argc"`
}

func (*rStartServiceARequest) Opnum() uint16 { return svcctl.OpnumRStartServiceA }

// rStartServiceAResponse carries the [out] parameters and return value of RStartServiceA.
type rStartServiceAResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RStartServiceA calls RStartServiceA (opnum 31) ([MS-SCMR] — verify the parameter
// modeling and status handling).
func RStartServiceA(rpc ndr.Invoker, hService msscmr.SC_RPC_HANDLE, argc ndr.DWORD, argv []msscmr.STRING_PTRSA) (err error) {
	req := &rStartServiceARequest{
		HService: hService,
		Argc:     argc,
		Argv:     argv,
	}
	var resp rStartServiceAResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RStartServiceA: %w", err)
		return
	}
	if uint32(resp.Status) != svcctl.StatusSuccess {
		err = fmt.Errorf("RStartServiceA failed: %s", svcctl.StatusString(uint32(resp.Status)))
	}
	return
}
