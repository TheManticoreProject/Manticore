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

// rOpenSCManager2Request carries the [in] parameters of ROpenSCManager2.
type rOpenSCManager2Request struct {
	DatabaseName  *ndr.WSTR `ndr:"unique"`
	DesiredAccess ndr.DWORD
}

func (*rOpenSCManager2Request) Opnum() uint16 { return svcctl.OpnumROpenSCManager2 }

// rOpenSCManager2Response carries the [out] parameters and return value of ROpenSCManager2.
type rOpenSCManager2Response struct {
	ScmHandle msscmr.LPSC_RPC_HANDLE
	Status    ndr.DWORD `ndr:"retval"`
}

// ROpenSCManager2 calls ROpenSCManager2 (opnum 64) ([MS-SCMR] — verify the parameter
// modeling and status handling).
func ROpenSCManager2(rpc ndr.Invoker, databaseName *ndr.WSTR, desiredAccess ndr.DWORD) (ScmHandle msscmr.LPSC_RPC_HANDLE, err error) {
	req := &rOpenSCManager2Request{
		DatabaseName:  databaseName,
		DesiredAccess: desiredAccess,
	}
	var resp rOpenSCManager2Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ROpenSCManager2: %w", err)
		return
	}
	ScmHandle = resp.ScmHandle
	if uint32(resp.Status) != svcctl.StatusSuccess {
		err = fmt.Errorf("ROpenSCManager2 failed: %s", svcctl.StatusString(uint32(resp.Status)))
	}
	return
}
