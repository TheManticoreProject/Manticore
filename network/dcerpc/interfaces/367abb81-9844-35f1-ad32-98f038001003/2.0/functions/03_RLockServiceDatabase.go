package functions

import (
	"fmt"

	svcctl "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/367abb81-9844-35f1-ad32-98f038001003/2.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/367abb81-9844-35f1-ad32-98f038001003/2.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rLockServiceDatabaseRequest carries the [in] parameters of RLockServiceDatabase.
type rLockServiceDatabaseRequest struct {
	HSCManager structures.SC_RPC_HANDLE
}

func (*rLockServiceDatabaseRequest) Opnum() uint16 { return svcctl.OpnumRLockServiceDatabase }

// rLockServiceDatabaseResponse carries the [out] parameters and return value of RLockServiceDatabase.
type rLockServiceDatabaseResponse struct {
	LpLock structures.LPSC_RPC_LOCK
	Status ndr.DWORD `ndr:"retval"`
}

// RLockServiceDatabase calls RLockServiceDatabase (opnum 3) ([MS-SCMR] — verify the parameter
// modeling and status handling).
func RLockServiceDatabase(rpc ndr.Invoker, hSCManager structures.SC_RPC_HANDLE) (LpLock structures.LPSC_RPC_LOCK, err error) {
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
