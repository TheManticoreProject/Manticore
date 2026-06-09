package functions

import (
	"fmt"

	svcctl "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/367abb81-9844-35f1-ad32-98f038001003/2.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/367abb81-9844-35f1-ad32-98f038001003/2.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rCreateServiceARequest carries the [in] parameters of RCreateServiceA.
type rCreateServiceARequest struct {
	HSCManager         structures.SC_RPC_HANDLE
	LpServiceName      ndr.STR
	LpDisplayName      *ndr.STR `ndr:"unique"`
	DwDesiredAccess    ndr.DWORD
	DwServiceType      ndr.DWORD
	DwStartType        ndr.DWORD
	DwErrorControl     ndr.DWORD
	LpBinaryPathName   ndr.STR
	LpLoadOrderGroup   *ndr.STR   `ndr:"unique"`
	LpdwTagId          *ndr.DWORD `ndr:"unique"`
	LpDependencies     []uint8    `ndr:"ref,size_is=DwDependSize"`
	DwDependSize       ndr.DWORD
	LpServiceStartName *ndr.STR `ndr:"unique"`
	LpPassword         []uint8  `ndr:"ref,size_is=DwPwSize"`
	DwPwSize           ndr.DWORD
}

func (*rCreateServiceARequest) Opnum() uint16 { return svcctl.OpnumRCreateServiceA }

// rCreateServiceAResponse carries the [out] parameters and return value of RCreateServiceA.
type rCreateServiceAResponse struct {
	LpdwTagId       *ndr.DWORD `ndr:"unique"`
	LpServiceHandle structures.LPSC_RPC_HANDLE
	Status          ndr.DWORD `ndr:"retval"`
}

// RCreateServiceA calls RCreateServiceA (opnum 24) ([MS-SCMR] — verify the parameter
// modeling and status handling).
func RCreateServiceA(rpc ndr.Invoker, hSCManager structures.SC_RPC_HANDLE, lpServiceName ndr.STR, lpDisplayName *ndr.STR, dwDesiredAccess ndr.DWORD, dwServiceType ndr.DWORD, dwStartType ndr.DWORD, dwErrorControl ndr.DWORD, lpBinaryPathName ndr.STR, lpLoadOrderGroup *ndr.STR, lpdwTagId *ndr.DWORD, lpDependencies []uint8, dwDependSize ndr.DWORD, lpServiceStartName *ndr.STR, lpPassword []uint8, dwPwSize ndr.DWORD) (LpdwTagId *ndr.DWORD, LpServiceHandle structures.LPSC_RPC_HANDLE, err error) {
	req := &rCreateServiceARequest{
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
	var resp rCreateServiceAResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RCreateServiceA: %w", err)
		return
	}
	LpdwTagId = resp.LpdwTagId
	LpServiceHandle = resp.LpServiceHandle
	if uint32(resp.Status) != svcctl.StatusSuccess {
		err = fmt.Errorf("RCreateServiceA failed: %s", svcctl.StatusString(uint32(resp.Status)))
	}
	return
}
