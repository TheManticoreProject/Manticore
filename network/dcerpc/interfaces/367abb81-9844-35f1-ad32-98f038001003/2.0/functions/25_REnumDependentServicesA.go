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

// rEnumDependentServicesARequest carries the [in] parameters of REnumDependentServicesA.
type rEnumDependentServicesARequest struct {
	HService       msscmr.SC_RPC_HANDLE
	DwServiceState ndr.DWORD
	CbBufSize      ndr.DWORD
}

func (*rEnumDependentServicesARequest) Opnum() uint16 { return svcctl.OpnumREnumDependentServicesA }

// rEnumDependentServicesAResponse carries the [out] parameters and return value of REnumDependentServicesA.
type rEnumDependentServicesAResponse struct {
	LpServices         []uint8 `ndr:"ref,size_is=CbBufSize"`
	PcbBytesNeeded     msscmr.LPBOUNDED_DWORD_256K
	LpServicesReturned msscmr.LPBOUNDED_DWORD_256K
	Status             ndr.DWORD `ndr:"retval"`
}

// REnumDependentServicesA calls REnumDependentServicesA (opnum 25) ([MS-SCMR] — verify the parameter
// modeling and status handling).
func REnumDependentServicesA(rpc ndr.Invoker, hService msscmr.SC_RPC_HANDLE, dwServiceState ndr.DWORD, cbBufSize ndr.DWORD) (LpServices []uint8, PcbBytesNeeded msscmr.LPBOUNDED_DWORD_256K, LpServicesReturned msscmr.LPBOUNDED_DWORD_256K, err error) {
	req := &rEnumDependentServicesARequest{
		HService:       hService,
		DwServiceState: dwServiceState,
		CbBufSize:      cbBufSize,
	}
	var resp rEnumDependentServicesAResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("REnumDependentServicesA: %w", err)
		return
	}
	LpServices = resp.LpServices
	PcbBytesNeeded = resp.PcbBytesNeeded
	LpServicesReturned = resp.LpServicesReturned
	if uint32(resp.Status) != svcctl.StatusSuccess {
		err = fmt.Errorf("REnumDependentServicesA failed: %s", svcctl.StatusString(uint32(resp.Status)))
	}
	return
}
