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

// rChangeServiceConfigWRequest carries the [in] parameters of RChangeServiceConfigW.
type rChangeServiceConfigWRequest struct {
	HService           msscmr.SC_RPC_HANDLE
	DwServiceType      ndr.DWORD
	DwStartType        ndr.DWORD
	DwErrorControl     ndr.DWORD
	LpBinaryPathName   *ndr.WSTR  `ndr:"unique"`
	LpLoadOrderGroup   *ndr.WSTR  `ndr:"unique"`
	LpdwTagId          *ndr.DWORD `ndr:"unique"`
	LpDependencies     []uint8    `ndr:"ref,size_is=DwDependSize"`
	DwDependSize       ndr.DWORD
	LpServiceStartName *ndr.WSTR `ndr:"unique"`
	LpPassword         []uint8   `ndr:"ref,size_is=DwPwSize"`
	DwPwSize           ndr.DWORD
	LpDisplayName      *ndr.WSTR `ndr:"unique"`
}

func (*rChangeServiceConfigWRequest) Opnum() uint16 { return svcctl.OpnumRChangeServiceConfigW }

// rChangeServiceConfigWResponse carries the [out] parameters and return value of RChangeServiceConfigW.
type rChangeServiceConfigWResponse struct {
	LpdwTagId *ndr.DWORD `ndr:"unique"`
	Status    ndr.DWORD  `ndr:"retval"`
}

// RChangeServiceConfigW calls RChangeServiceConfigW (opnum 11) ([MS-SCMR] — verify the parameter
// modeling and status handling).
func RChangeServiceConfigW(rpc ndr.Invoker, hService msscmr.SC_RPC_HANDLE, dwServiceType ndr.DWORD, dwStartType ndr.DWORD, dwErrorControl ndr.DWORD, lpBinaryPathName *ndr.WSTR, lpLoadOrderGroup *ndr.WSTR, lpdwTagId *ndr.DWORD, lpDependencies []uint8, dwDependSize ndr.DWORD, lpServiceStartName *ndr.WSTR, lpPassword []uint8, dwPwSize ndr.DWORD, lpDisplayName *ndr.WSTR) (LpdwTagId *ndr.DWORD, err error) {
	req := &rChangeServiceConfigWRequest{
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
	var resp rChangeServiceConfigWResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RChangeServiceConfigW: %w", err)
		return
	}
	LpdwTagId = resp.LpdwTagId
	if uint32(resp.Status) != svcctl.StatusSuccess {
		err = fmt.Errorf("RChangeServiceConfigW failed: %s", svcctl.StatusString(uint32(resp.Status)))
	}
	return
}
