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

// rOpenServiceWRequest carries the [in] parameters of ROpenServiceW.
type rOpenServiceWRequest struct {
	HSCManager      msscmr.SC_RPC_HANDLE
	LpServiceName   ndr.WSTR
	DwDesiredAccess ndr.DWORD
}

func (*rOpenServiceWRequest) Opnum() uint16 { return svcctl.OpnumROpenServiceW }

// rOpenServiceWResponse carries the [out] parameters and return value of ROpenServiceW.
type rOpenServiceWResponse struct {
	LpServiceHandle msscmr.LPSC_RPC_HANDLE
	Status          ndr.DWORD `ndr:"retval"`
}

// ROpenServiceW calls ROpenServiceW (opnum 16) ([MS-SCMR] — verify the parameter
// modeling and status handling).
func ROpenServiceW(rpc ndr.Invoker, hSCManager msscmr.SC_RPC_HANDLE, lpServiceName ndr.WSTR, dwDesiredAccess ndr.DWORD) (LpServiceHandle msscmr.LPSC_RPC_HANDLE, err error) {
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
