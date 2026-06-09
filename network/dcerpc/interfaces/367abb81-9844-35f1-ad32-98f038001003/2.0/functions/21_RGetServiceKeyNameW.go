package functions

import (
	"fmt"

	svcctl "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/367abb81-9844-35f1-ad32-98f038001003/2.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/367abb81-9844-35f1-ad32-98f038001003/2.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rGetServiceKeyNameWRequest carries the [in] parameters of RGetServiceKeyNameW.
type rGetServiceKeyNameWRequest struct {
	HSCManager    structures.SC_RPC_HANDLE
	LpDisplayName ndr.WSTR
	LpcchBuffer   ndr.DWORD
}

func (*rGetServiceKeyNameWRequest) Opnum() uint16 { return svcctl.OpnumRGetServiceKeyNameW }

// rGetServiceKeyNameWResponse carries the [out] parameters and return value of RGetServiceKeyNameW.
type rGetServiceKeyNameWResponse struct {
	LpServiceName ndr.WSTR
	LpcchBuffer   ndr.DWORD
	Status        ndr.DWORD `ndr:"retval"`
}

// RGetServiceKeyNameW calls RGetServiceKeyNameW (opnum 21) ([MS-SCMR] — verify the parameter
// modeling and status handling).
func RGetServiceKeyNameW(rpc ndr.Invoker, hSCManager structures.SC_RPC_HANDLE, lpDisplayName ndr.WSTR, lpcchBuffer ndr.DWORD) (LpServiceName ndr.WSTR, LpcchBuffer ndr.DWORD, err error) {
	req := &rGetServiceKeyNameWRequest{
		HSCManager:    hSCManager,
		LpDisplayName: lpDisplayName,
		LpcchBuffer:   lpcchBuffer,
	}
	var resp rGetServiceKeyNameWResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RGetServiceKeyNameW: %w", err)
		return
	}
	LpServiceName = resp.LpServiceName
	LpcchBuffer = resp.LpcchBuffer
	if uint32(resp.Status) != svcctl.StatusSuccess {
		err = fmt.Errorf("RGetServiceKeyNameW failed: %s", svcctl.StatusString(uint32(resp.Status)))
	}
	return
}
