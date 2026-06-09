package functions

import (
	"fmt"

	svcctl "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/367abb81-9844-35f1-ad32-98f038001003/2.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/367abb81-9844-35f1-ad32-98f038001003/2.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rOpenServiceWRequest carries the [in] parameters of ROpenServiceW.
type rOpenServiceWRequest struct {
	HSCManager      structures.SC_RPC_HANDLE
	LpServiceName   ndr.WSTR
	DwDesiredAccess ndr.DWORD
}

func (*rOpenServiceWRequest) Opnum() uint16 { return svcctl.OpnumROpenServiceW }

// rOpenServiceWResponse carries the [out] parameters and return value of ROpenServiceW.
type rOpenServiceWResponse struct {
	LpServiceHandle structures.LPSC_RPC_HANDLE
	Status          ndr.DWORD `ndr:"retval"`
}

// ROpenServiceW calls ROpenServiceW (opnum 16) ([MS-SCMR] — verify the parameter
// modeling and status handling).
func ROpenServiceW(rpc ndr.Invoker, hSCManager structures.SC_RPC_HANDLE, lpServiceName ndr.WSTR, dwDesiredAccess ndr.DWORD) (LpServiceHandle structures.LPSC_RPC_HANDLE, err error) {
	req := &rOpenServiceWRequest{
		HSCManager:      hSCManager,
		LpServiceName:   lpServiceName,
		DwDesiredAccess: dwDesiredAccess,
	}
	var resp rOpenServiceWResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ROpenServiceW: %w", err)
		return
	}
	LpServiceHandle = resp.LpServiceHandle
	if uint32(resp.Status) != svcctl.StatusSuccess {
		err = fmt.Errorf("ROpenServiceW failed: %s", svcctl.StatusString(uint32(resp.Status)))
	}
	return
}
