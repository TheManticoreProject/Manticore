package functions

import (
	"fmt"

	svcctl "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/367abb81-9844-35f1-ad32-98f038001003/2.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msscmr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-scmr"
)

// rOpenServiceARequest carries the [in] parameters of ROpenServiceA.
type rOpenServiceARequest struct {
	HSCManager      msscmr.SC_RPC_HANDLE
	LpServiceName   ndr.STR
	DwDesiredAccess ndr.DWORD
}

func (*rOpenServiceARequest) Opnum() uint16 { return svcctl.OpnumROpenServiceA }

// rOpenServiceAResponse carries the [out] parameters and return value of ROpenServiceA.
type rOpenServiceAResponse struct {
	LpServiceHandle msscmr.LPSC_RPC_HANDLE
	Status          ndr.DWORD `ndr:"retval"`
}

// ROpenServiceA calls ROpenServiceA (opnum 28) ([MS-SCMR] — verify the parameter
// modeling and status handling).
func ROpenServiceA(rpc ndr.Invoker, hSCManager msscmr.SC_RPC_HANDLE, lpServiceName ndr.STR, dwDesiredAccess ndr.DWORD) (LpServiceHandle msscmr.LPSC_RPC_HANDLE, err error) {
	req := &rOpenServiceARequest{
		HSCManager:      hSCManager,
		LpServiceName:   lpServiceName,
		DwDesiredAccess: dwDesiredAccess,
	}
	var resp rOpenServiceAResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ROpenServiceA: %w", err)
		return
	}
	LpServiceHandle = resp.LpServiceHandle
	if uint32(resp.Status) != svcctl.StatusSuccess {
		err = fmt.Errorf("ROpenServiceA failed: %s", svcctl.StatusString(uint32(resp.Status)))
	}
	return
}
