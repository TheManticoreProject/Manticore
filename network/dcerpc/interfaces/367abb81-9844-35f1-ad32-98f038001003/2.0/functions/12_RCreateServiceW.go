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

// rCreateServiceWRequest carries the [in] parameters of RCreateServiceW.
type rCreateServiceWRequest struct {
	HSCManager         msscmr.SC_RPC_HANDLE
	LpServiceName      ndr.WSTR
	LpDisplayName      *ndr.WSTR `ndr:"unique"`
	DwDesiredAccess    ndr.DWORD
	DwServiceType      ndr.DWORD
	DwStartType        ndr.DWORD
	DwErrorControl     ndr.DWORD
	LpBinaryPathName   ndr.WSTR
	LpLoadOrderGroup   *ndr.WSTR  `ndr:"unique"`
	LpdwTagId          *ndr.DWORD `ndr:"unique"`
	LpDependencies     []uint8    `ndr:"ref,size_is=DwDependSize"`
	DwDependSize       ndr.DWORD
	LpServiceStartName *ndr.WSTR `ndr:"unique"`
	LpPassword         []uint8   `ndr:"ref,size_is=DwPwSize"`
	DwPwSize           ndr.DWORD
}

func (*rCreateServiceWRequest) Opnum() uint16 { return svcctl.OpnumRCreateServiceW }

// rCreateServiceWResponse carries the [out] parameters and return value of RCreateServiceW.
type rCreateServiceWResponse struct {
	LpdwTagId       *ndr.DWORD `ndr:"unique"`
	LpServiceHandle msscmr.LPSC_RPC_HANDLE
	Status          ndr.DWORD `ndr:"retval"`
}

// RCreateServiceW calls RCreateServiceW (opnum 12) ([MS-SCMR] — verify the parameter
// modeling and status handling).
func RCreateServiceW(rpc ndr.Invoker, hSCManager msscmr.SC_RPC_HANDLE, lpServiceName ndr.WSTR, lpDisplayName *ndr.WSTR, dwDesiredAccess ndr.DWORD, dwServiceType ndr.DWORD, dwStartType ndr.DWORD, dwErrorControl ndr.DWORD, lpBinaryPathName ndr.WSTR, lpLoadOrderGroup *ndr.WSTR, lpdwTagId *ndr.DWORD, lpDependencies []uint8, dwDependSize ndr.DWORD, lpServiceStartName *ndr.WSTR, lpPassword []uint8, dwPwSize ndr.DWORD) (LpdwTagId *ndr.DWORD, LpServiceHandle msscmr.LPSC_RPC_HANDLE, err error) {
	req := &rCreateServiceWRequest{
		HSCManager:         hSCManager,
		LpServiceName:      lpServiceName,
		LpDisplayName:      lpDisplayName,
		DwDesiredAccess:    dwDesiredAccess,
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
	}
	var resp rCreateServiceWResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RCreateServiceW: %w", err)
		return
	}
	LpdwTagId = resp.LpdwTagId
	LpServiceHandle = resp.LpServiceHandle
	if uint32(resp.Status) != svcctl.StatusSuccess {
		err = fmt.Errorf("RCreateServiceW failed: %s", svcctl.StatusString(uint32(resp.Status)))
	}
	return
}
