package functions

import (
	"fmt"

	svcctl "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/367abb81-9844-35f1-ad32-98f038001003/2.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msscmr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-scmr"
)

// rGetServiceDisplayNameARequest carries the [in] parameters of RGetServiceDisplayNameA.
type rGetServiceDisplayNameARequest struct {
	HSCManager    msscmr.SC_RPC_HANDLE
	LpServiceName ndr.STR
	LpcchBuffer   msscmr.LPBOUNDED_DWORD_4K
}

func (*rGetServiceDisplayNameARequest) Opnum() uint16 { return svcctl.OpnumRGetServiceDisplayNameA }

// rGetServiceDisplayNameAResponse carries the [out] parameters and return value of RGetServiceDisplayNameA.
type rGetServiceDisplayNameAResponse struct {
	LpDisplayName ndr.STR
	LpcchBuffer   msscmr.LPBOUNDED_DWORD_4K
	Status        ndr.DWORD `ndr:"retval"`
}

// RGetServiceDisplayNameA calls RGetServiceDisplayNameA (opnum 32) ([MS-SCMR] — verify the parameter
// modeling and status handling).
func RGetServiceDisplayNameA(rpc ndr.Invoker, hSCManager msscmr.SC_RPC_HANDLE, lpServiceName ndr.STR, lpcchBuffer msscmr.LPBOUNDED_DWORD_4K) (LpDisplayName ndr.STR, LpcchBuffer msscmr.LPBOUNDED_DWORD_4K, err error) {
	req := &rGetServiceDisplayNameARequest{
		HSCManager:    hSCManager,
		LpServiceName: lpServiceName,
		LpcchBuffer:   lpcchBuffer,
	}
	var resp rGetServiceDisplayNameAResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RGetServiceDisplayNameA: %w", err)
		return
	}
	LpDisplayName = resp.LpDisplayName
	LpcchBuffer = resp.LpcchBuffer
	if uint32(resp.Status) != svcctl.StatusSuccess {
		err = fmt.Errorf("RGetServiceDisplayNameA failed: %s", svcctl.StatusString(uint32(resp.Status)))
	}
	return
}
