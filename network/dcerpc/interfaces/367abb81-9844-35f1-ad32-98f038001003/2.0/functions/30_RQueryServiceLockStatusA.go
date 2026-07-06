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

// rQueryServiceLockStatusARequest carries the [in] parameters of RQueryServiceLockStatusA.
type rQueryServiceLockStatusARequest struct {
	HSCManager msscmr.SC_RPC_HANDLE
	CbBufSize  ndr.DWORD
}

func (*rQueryServiceLockStatusARequest) Opnum() uint16 { return svcctl.OpnumRQueryServiceLockStatusA }

// rQueryServiceLockStatusAResponse carries the [out] parameters and return value of RQueryServiceLockStatusA.
type rQueryServiceLockStatusAResponse struct {
	LpLockStatus   msscmr.QUERY_SERVICE_LOCK_STATUSA
	PcbBytesNeeded msscmr.LPBOUNDED_DWORD_4K
	Status         ndr.DWORD `ndr:"retval"`
}

// RQueryServiceLockStatusA calls RQueryServiceLockStatusA (opnum 30) ([MS-SCMR] — verify the parameter
// modeling and status handling).
func RQueryServiceLockStatusA(rpc ndr.Invoker, hSCManager msscmr.SC_RPC_HANDLE, cbBufSize ndr.DWORD) (LpLockStatus msscmr.QUERY_SERVICE_LOCK_STATUSA, PcbBytesNeeded msscmr.LPBOUNDED_DWORD_4K, err error) {
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
