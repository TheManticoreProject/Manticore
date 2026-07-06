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

// rOpenSCManagerARequest carries the [in] parameters of ROpenSCManagerA.
type rOpenSCManagerARequest struct {
	LpMachineName   *ndr.STR `ndr:"unique"`
	LpDatabaseName  *ndr.STR `ndr:"unique"`
	DwDesiredAccess ndr.DWORD
}

func (*rOpenSCManagerARequest) Opnum() uint16 { return svcctl.OpnumROpenSCManagerA }

// rOpenSCManagerAResponse carries the [out] parameters and return value of ROpenSCManagerA.
type rOpenSCManagerAResponse struct {
	LpScHandle msscmr.LPSC_RPC_HANDLE
	Status     ndr.DWORD `ndr:"retval"`
}

// ROpenSCManagerA calls ROpenSCManagerA (opnum 27) ([MS-SCMR] — verify the parameter
// modeling and status handling).
func ROpenSCManagerA(rpc ndr.Invoker, lpMachineName *ndr.STR, lpDatabaseName *ndr.STR, dwDesiredAccess ndr.DWORD) (LpScHandle msscmr.LPSC_RPC_HANDLE, err error) {
	req := &rOpenSCManagerARequest{
		LpMachineName:   lpMachineName,
		LpDatabaseName:  lpDatabaseName,
		DwDesiredAccess: dwDesiredAccess,
	}
	var resp rOpenSCManagerAResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ROpenSCManagerA: %w", err)
		return
	}
	LpScHandle = resp.LpScHandle
	if uint32(resp.Status) != svcctl.StatusSuccess {
		err = fmt.Errorf("ROpenSCManagerA failed: %s", svcctl.StatusString(uint32(resp.Status)))
	}
	return
}
