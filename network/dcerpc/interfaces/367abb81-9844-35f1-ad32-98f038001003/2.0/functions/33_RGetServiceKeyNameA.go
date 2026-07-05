package functions

import (
	"fmt"

	svcctl "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/367abb81-9844-35f1-ad32-98f038001003/2.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msscmr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-scmr"
)

// rGetServiceKeyNameARequest carries the [in] parameters of RGetServiceKeyNameA.
type rGetServiceKeyNameARequest struct {
	HSCManager    msscmr.SC_RPC_HANDLE
	LpDisplayName ndr.STR
	LpcchBuffer   msscmr.LPBOUNDED_DWORD_4K
}

func (*rGetServiceKeyNameARequest) Opnum() uint16 { return svcctl.OpnumRGetServiceKeyNameA }

// rGetServiceKeyNameAResponse carries the [out] parameters and return value of RGetServiceKeyNameA.
type rGetServiceKeyNameAResponse struct {
	LpKeyName   ndr.STR
	LpcchBuffer msscmr.LPBOUNDED_DWORD_4K
	Status      ndr.DWORD `ndr:"retval"`
}

// RGetServiceKeyNameA calls RGetServiceKeyNameA (opnum 33) ([MS-SCMR] — verify the parameter
// modeling and status handling).
func RGetServiceKeyNameA(rpc ndr.Invoker, hSCManager msscmr.SC_RPC_HANDLE, lpDisplayName ndr.STR, lpcchBuffer msscmr.LPBOUNDED_DWORD_4K) (LpKeyName ndr.STR, LpcchBuffer msscmr.LPBOUNDED_DWORD_4K, err error) {
	req := &rGetServiceKeyNameARequest{
		HSCManager:    hSCManager,
		LpDisplayName: lpDisplayName,
		LpcchBuffer:   lpcchBuffer,
	}
	var resp rGetServiceKeyNameAResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RGetServiceKeyNameA: %w", err)
		return
	}
	LpKeyName = resp.LpKeyName
	LpcchBuffer = resp.LpcchBuffer
	if uint32(resp.Status) != svcctl.StatusSuccess {
		err = fmt.Errorf("RGetServiceKeyNameA failed: %s", svcctl.StatusString(uint32(resp.Status)))
	}
	return
}
