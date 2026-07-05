package functions

import (
	"fmt"

	W32Time "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/8fb6d884-2388-11d0-8c35-00c04fda2795/4.1"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// w32TimeLogRequest carries the [in] parameters of W32TimeLog.
type w32TimeLogRequest struct {
}

func (*w32TimeLogRequest) Opnum() uint16 { return W32Time.OpnumW32TimeLog }

// w32TimeLogResponse carries the [out] parameters and return value of W32TimeLog.
type w32TimeLogResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// W32TimeLog calls W32TimeLog (opnum 7) ([MS-W32T] section 3.2.4).
func W32TimeLog(rpc ndr.Invoker) (err error) {
	req := &w32TimeLogRequest{}
	var resp w32TimeLogResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("W32TimeLog: %w", err)
		return
	}
	if uint32(resp.Status) != W32Time.StatusSuccess {
		err = fmt.Errorf("W32TimeLog failed: %s", W32Time.StatusString(uint32(resp.Status)))
	}
	return
}
