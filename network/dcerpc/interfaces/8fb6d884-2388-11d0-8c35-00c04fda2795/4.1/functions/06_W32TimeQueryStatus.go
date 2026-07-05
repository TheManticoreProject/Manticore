package functions

import (
	"fmt"

	W32Time "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/8fb6d884-2388-11d0-8c35-00c04fda2795/4.1"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msw32t "github.com/TheManticoreProject/Manticore/windows/protocols/ms-w32t"
)

// w32TimeQueryStatusRequest carries the [in] parameters of W32TimeQueryStatus.
type w32TimeQueryStatusRequest struct {
}

func (*w32TimeQueryStatusRequest) Opnum() uint16 { return W32Time.OpnumW32TimeQueryStatus }

// w32TimeQueryStatusResponse carries the [out] parameters and return value of W32TimeQueryStatus.
type w32TimeQueryStatusResponse struct {
	PStatusInfo *msw32t.W32TIME_STATUS_INFO `ndr:"unique"`
	Status      ndr.DWORD                   `ndr:"retval"`
}

// W32TimeQueryStatus calls W32TimeQueryStatus (opnum 6) ([MS-W32T] section 3.2.4).
func W32TimeQueryStatus(rpc ndr.Invoker) (PStatusInfo *msw32t.W32TIME_STATUS_INFO, err error) {
	req := &w32TimeQueryStatusRequest{}
	var resp w32TimeQueryStatusResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("W32TimeQueryStatus: %w", err)
		return
	}
	PStatusInfo = resp.PStatusInfo
	if uint32(resp.Status) != W32Time.StatusSuccess {
		err = fmt.Errorf("W32TimeQueryStatus failed: %s", W32Time.StatusString(uint32(resp.Status)))
	}
	return
}
