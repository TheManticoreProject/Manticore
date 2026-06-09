package functions

import (
	"fmt"

	svcctl "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/367abb81-9844-35f1-ad32-98f038001003/2.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/367abb81-9844-35f1-ad32-98f038001003/2.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rCreateWowServiceRequest carries the [in] parameters of RCreateWowService.
type rCreateWowServiceRequest struct {
	HSCManager         structures.SC_RPC_HANDLE
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
	DwServiceWowType   uint16
}

func (*rCreateWowServiceRequest) Opnum() uint16 { return svcctl.OpnumRCreateWowService }

// rCreateWowServiceResponse carries the [out] parameters and return value of RCreateWowService.
type rCreateWowServiceResponse struct {
	LpdwTagId       *ndr.DWORD `ndr:"unique"`
	LpServiceHandle structures.LPSC_RPC_HANDLE
	Status          ndr.DWORD `ndr:"retval"`
}

// RCreateWowService calls RCreateWowService (opnum 60) ([MS-SCMR] — verify the parameter
// modeling and status handling).
func RCreateWowService(rpc ndr.Invoker, hSCManager structures.SC_RPC_HANDLE, lpServiceName ndr.WSTR, lpDisplayName *ndr.WSTR, dwDesiredAccess ndr.DWORD, dwServiceType ndr.DWORD, dwStartType ndr.DWORD, dwErrorControl ndr.DWORD, lpBinaryPathName ndr.WSTR, lpLoadOrderGroup *ndr.WSTR, lpdwTagId *ndr.DWORD, lpDependencies []uint8, dwDependSize ndr.DWORD, lpServiceStartName *ndr.WSTR, lpPassword []uint8, dwPwSize ndr.DWORD, dwServiceWowType uint16) (LpdwTagId *ndr.DWORD, LpServiceHandle structures.LPSC_RPC_HANDLE, err error) {
	req := &rCreateWowServiceRequest{
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
		DwServiceWowType:   dwServiceWowType,
	}
	var resp rCreateWowServiceResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RCreateWowService: %w", err)
		return
	}
	LpdwTagId = resp.LpdwTagId
	LpServiceHandle = resp.LpServiceHandle
	if uint32(resp.Status) != svcctl.StatusSuccess {
		err = fmt.Errorf("RCreateWowService failed: %s", svcctl.StatusString(uint32(resp.Status)))
	}
	return
}
