package functions

import (
	"fmt"

	svcctl "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/367abb81-9844-35f1-ad32-98f038001003/2.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/367abb81-9844-35f1-ad32-98f038001003/2.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rQueryServiceLockStatusARequest carries the [in] parameters of RQueryServiceLockStatusA.
type rQueryServiceLockStatusARequest struct {
	HSCManager structures.SC_RPC_HANDLE
	CbBufSize  ndr.DWORD
}

func (*rQueryServiceLockStatusARequest) Opnum() uint16 { return svcctl.OpnumRQueryServiceLockStatusA }

// rQueryServiceLockStatusAResponse carries the [out] parameters and return value of RQueryServiceLockStatusA.
type rQueryServiceLockStatusAResponse struct {
	LpLockStatus   structures.QUERY_SERVICE_LOCK_STATUSA
	PcbBytesNeeded structures.LPBOUNDED_DWORD_4K
	Status         ndr.DWORD `ndr:"retval"`
}

// RQueryServiceLockStatusA calls RQueryServiceLockStatusA (opnum 30) ([MS-SCMR] — verify the parameter
// modeling and status handling).
func RQueryServiceLockStatusA(rpc ndr.Invoker, hSCManager structures.SC_RPC_HANDLE, cbBufSize ndr.DWORD) (LpLockStatus structures.QUERY_SERVICE_LOCK_STATUSA, PcbBytesNeeded structures.LPBOUNDED_DWORD_4K, err error) {
	req := &rQueryServiceLockStatusARequest{
		HSCManager: hSCManager,
		CbBufSize:  cbBufSize,
	}
	var resp rQueryServiceLockStatusAResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RQueryServiceLockStatusA: %w", err)
		return
	}
	LpLockStatus = resp.LpLockStatus
	PcbBytesNeeded = resp.PcbBytesNeeded
	if uint32(resp.Status) != svcctl.StatusSuccess {
		err = fmt.Errorf("RQueryServiceLockStatusA failed: %s", svcctl.StatusString(uint32(resp.Status)))
	}
	return
}
