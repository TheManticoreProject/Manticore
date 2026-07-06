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

// rQueryServiceLockStatusWRequest carries the [in] parameters of RQueryServiceLockStatusW.
type rQueryServiceLockStatusWRequest struct {
	HSCManager msscmr.SC_RPC_HANDLE
	CbBufSize  ndr.DWORD
}

func (*rQueryServiceLockStatusWRequest) Opnum() uint16 { return svcctl.OpnumRQueryServiceLockStatusW }

// rQueryServiceLockStatusWResponse carries the [out] parameters and return value of RQueryServiceLockStatusW.
type rQueryServiceLockStatusWResponse struct {
	LpLockStatus   msscmr.QUERY_SERVICE_LOCK_STATUSW
	PcbBytesNeeded msscmr.LPBOUNDED_DWORD_4K
	Status         ndr.DWORD `ndr:"retval"`
}

// RQueryServiceLockStatusW calls RQueryServiceLockStatusW (opnum 18) ([MS-SCMR] — verify the parameter
// modeling and status handling).
func RQueryServiceLockStatusW(rpc ndr.Invoker, hSCManager msscmr.SC_RPC_HANDLE, cbBufSize ndr.DWORD) (LpLockStatus msscmr.QUERY_SERVICE_LOCK_STATUSW, PcbBytesNeeded msscmr.LPBOUNDED_DWORD_4K, err error) {
	req := &rQueryServiceLockStatusWRequest{
		HSCManager: hSCManager,
		CbBufSize:  cbBufSize,
	}
	var resp rQueryServiceLockStatusWResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RQueryServiceLockStatusW: %w", err)
		return
	}
	LpLockStatus = resp.LpLockStatus
	PcbBytesNeeded = resp.PcbBytesNeeded
	if uint32(resp.Status) != svcctl.StatusSuccess {
		err = fmt.Errorf("RQueryServiceLockStatusW failed: %s", svcctl.StatusString(uint32(resp.Status)))
	}
	return
}
