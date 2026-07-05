package functions

import (
	"fmt"

	W32Time "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/8fb6d884-2388-11d0-8c35-00c04fda2795/4.1"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// w32TimeSyncRequest carries the [in] parameters of W32TimeSync.
type w32TimeSyncRequest struct {
	UWait   ndr.DWORD
	UlFlags ndr.DWORD
}

func (*w32TimeSyncRequest) Opnum() uint16 { return W32Time.OpnumW32TimeSync }

// w32TimeSyncResponse carries the [out] parameters and return value of W32TimeSync.
type w32TimeSyncResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// W32TimeSync calls W32TimeSync (opnum 0) ([MS-W32T] section 3.2.4).
func W32TimeSync(rpc ndr.Invoker, uWait ndr.DWORD, ulFlags ndr.DWORD) (err error) {
	req := &w32TimeSyncRequest{
		UWait:   uWait,
		UlFlags: ulFlags,
	}
	var resp w32TimeSyncResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("W32TimeSync: %w", err)
		return
	}
	if uint32(resp.Status) != W32Time.StatusSuccess {
		err = fmt.Errorf("W32TimeSync failed: %s", W32Time.StatusString(uint32(resp.Status)))
	}
	return
}
