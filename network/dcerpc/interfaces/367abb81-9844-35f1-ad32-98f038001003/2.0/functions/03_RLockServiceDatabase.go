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

// rLockServiceDatabaseRequest carries the [in] parameters of RLockServiceDatabase.
type rLockServiceDatabaseRequest struct {
	HSCManager msscmr.SC_RPC_HANDLE
}

func (*rLockServiceDatabaseRequest) Opnum() uint16 { return svcctl.OpnumRLockServiceDatabase }

// rLockServiceDatabaseResponse carries the [out] parameters and return value of RLockServiceDatabase.
type rLockServiceDatabaseResponse struct {
	LpLock msscmr.LPSC_RPC_LOCK
	Status ndr.DWORD `ndr:"retval"`
}

// RLockServiceDatabase calls RLockServiceDatabase (opnum 3) ([MS-SCMR] — verify the parameter
// modeling and status handling).
func RLockServiceDatabase(rpc ndr.Invoker, hSCManager msscmr.SC_RPC_HANDLE) (LpLock msscmr.LPSC_RPC_LOCK, err error) {
	req := &rLockServiceDatabaseRequest{
		HSCManager: hSCManager,
	}
	var resp rLockServiceDatabaseResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RLockServiceDatabase: %w", err)
		return
	}
	LpLock = resp.LpLock
	if uint32(resp.Status) != svcctl.StatusSuccess {
		err = fmt.Errorf("RLockServiceDatabase failed: %s", svcctl.StatusString(uint32(resp.Status)))
	}
	return
}
