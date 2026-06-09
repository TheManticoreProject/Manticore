package functions

import (
	"fmt"

	svcctl "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/367abb81-9844-35f1-ad32-98f038001003/2.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/367abb81-9844-35f1-ad32-98f038001003/2.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rUnlockServiceDatabaseRequest carries the [in] parameters of RUnlockServiceDatabase.
type rUnlockServiceDatabaseRequest struct {
	Lock structures.LPSC_RPC_LOCK
}

func (*rUnlockServiceDatabaseRequest) Opnum() uint16 { return svcctl.OpnumRUnlockServiceDatabase }

// rUnlockServiceDatabaseResponse carries the [out] parameters and return value of RUnlockServiceDatabase.
type rUnlockServiceDatabaseResponse struct {
	Lock   structures.LPSC_RPC_LOCK
	Status ndr.DWORD `ndr:"retval"`
}

// RUnlockServiceDatabase calls RUnlockServiceDatabase (opnum 8) ([MS-SCMR] — verify the parameter
// modeling and status handling).
func RUnlockServiceDatabase(rpc ndr.Invoker, lock structures.LPSC_RPC_LOCK) (Lock structures.LPSC_RPC_LOCK, err error) {
	req := &rUnlockServiceDatabaseRequest{
		Lock: lock,
	}
	var resp rUnlockServiceDatabaseResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RUnlockServiceDatabase: %w", err)
		return
	}
	Lock = resp.Lock
	if uint32(resp.Status) != svcctl.StatusSuccess {
		err = fmt.Errorf("RUnlockServiceDatabase failed: %s", svcctl.StatusString(uint32(resp.Status)))
	}
	return
}
