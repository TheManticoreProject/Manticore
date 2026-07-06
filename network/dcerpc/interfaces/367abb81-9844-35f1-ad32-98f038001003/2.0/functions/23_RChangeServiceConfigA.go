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

// rChangeServiceConfigARequest carries the [in] parameters of RChangeServiceConfigA.
type rChangeServiceConfigARequest struct {
	HService           msscmr.SC_RPC_HANDLE
	DwServiceType      ndr.DWORD
	DwStartType        ndr.DWORD
	DwErrorControl     ndr.DWORD
	LpBinaryPathName   *ndr.STR   `ndr:"unique"`
	LpLoadOrderGroup   *ndr.STR   `ndr:"unique"`
	LpdwTagId          *ndr.DWORD `ndr:"unique"`
	LpDependencies     []uint8    `ndr:"ref,size_is=DwDependSize"`
	DwDependSize       ndr.DWORD
	LpServiceStartName *ndr.STR `ndr:"unique"`
	LpPassword         []uint8  `ndr:"ref,size_is=DwPwSize"`
	DwPwSize           ndr.DWORD
	LpDisplayName      *ndr.STR `ndr:"unique"`
}

func (*rChangeServiceConfigARequest) Opnum() uint16 { return svcctl.OpnumRChangeServiceConfigA }

// rChangeServiceConfigAResponse carries the [out] parameters and return value of RChangeServiceConfigA.
type rChangeServiceConfigAResponse struct {
	LpdwTagId *ndr.DWORD `ndr:"unique"`
	Status    ndr.DWORD  `ndr:"retval"`
}

// RChangeServiceConfigA calls RChangeServiceConfigA (opnum 23) ([MS-SCMR] — verify the parameter
// modeling and status handling).
func RChangeServiceConfigA(rpc ndr.Invoker, hService msscmr.SC_RPC_HANDLE, dwServiceType ndr.DWORD, dwStartType ndr.DWORD, dwErrorControl ndr.DWORD, lpBinaryPathName *ndr.STR, lpLoadOrderGroup *ndr.STR, lpdwTagId *ndr.DWORD, lpDependencies []uint8, dwDependSize ndr.DWORD, lpServiceStartName *ndr.STR, lpPassword []uint8, dwPwSize ndr.DWORD, lpDisplayName *ndr.STR) (LpdwTagId *ndr.DWORD, err error) {
	req := &rChangeServiceConfigARequest{
		HService:           hService,
		DwServiceType:      dwServiceType,
		DwStartType:        dwStartType,
		DwErrorControl:     dwErrorControl,
		LpBinaryPathName:   lpBinaryPathName,
		LpLoadOrderGroup:   lpLoadOrderGroup,
		LpdwTagId:          lpdwTagId,
		LpDependencies:     lpDependencies,
		DwDependSize:       dwDependSize,
		LpServiceStartName: lpServiceStartName,
		LpPassword:         lpPassword,
		DwPwSize:           dwPwSize,
		LpDisplayName:      lpDisplayName,
	}
	var resp rChangeServiceConfigAResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RChangeServiceConfigA: %w", err)
		return
	}
	LpdwTagId = resp.LpdwTagId
	if uint32(resp.Status) != svcctl.StatusSuccess {
		err = fmt.Errorf("RChangeServiceConfigA failed: %s", svcctl.StatusString(uint32(resp.Status)))
	}
	return
}
