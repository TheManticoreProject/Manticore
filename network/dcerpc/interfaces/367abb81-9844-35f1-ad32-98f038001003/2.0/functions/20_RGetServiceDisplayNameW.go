package functions

import (
	"fmt"

	svcctl "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/367abb81-9844-35f1-ad32-98f038001003/2.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/367abb81-9844-35f1-ad32-98f038001003/2.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rGetServiceDisplayNameWRequest carries the [in] parameters of RGetServiceDisplayNameW.
type rGetServiceDisplayNameWRequest struct {
	HSCManager    structures.SC_RPC_HANDLE
	LpServiceName ndr.WSTR
	LpcchBuffer   ndr.DWORD
}

func (*rGetServiceDisplayNameWRequest) Opnum() uint16 { return svcctl.OpnumRGetServiceDisplayNameW }

// rGetServiceDisplayNameWResponse carries the [out] parameters and return value of RGetServiceDisplayNameW.
type rGetServiceDisplayNameWResponse struct {
	LpDisplayName ndr.WSTR
	LpcchBuffer   ndr.DWORD
	Status        ndr.DWORD `ndr:"retval"`
}

// RGetServiceDisplayNameW calls RGetServiceDisplayNameW (opnum 20) ([MS-SCMR] — verify the parameter
// modeling and status handling).
func RGetServiceDisplayNameW(rpc ndr.Invoker, hSCManager structures.SC_RPC_HANDLE, lpServiceName ndr.WSTR, lpcchBuffer ndr.DWORD) (LpDisplayName ndr.WSTR, LpcchBuffer ndr.DWORD, err error) {
	req := &rGetServiceDisplayNameWRequest{
		HSCManager:    hSCManager,
		LpServiceName: lpServiceName,
		LpcchBuffer:   lpcchBuffer,
	}
	var resp rGetServiceDisplayNameWResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RGetServiceDisplayNameW: %w", err)
		return
	}
	LpDisplayName = resp.LpDisplayName
	LpcchBuffer = resp.LpcchBuffer
	if uint32(resp.Status) != svcctl.StatusSuccess {
		err = fmt.Errorf("RGetServiceDisplayNameW failed: %s", svcctl.StatusString(uint32(resp.Status)))
	}
	return
}
