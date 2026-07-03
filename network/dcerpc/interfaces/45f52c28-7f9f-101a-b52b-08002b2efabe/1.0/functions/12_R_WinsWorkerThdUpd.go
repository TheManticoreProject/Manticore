package functions

import (
	"fmt"

	winsif "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/45f52c28-7f9f-101a-b52b-08002b2efabe/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// r_WinsWorkerThdUpdRequest carries the [in] parameters of R_WinsWorkerThdUpd.
type r_WinsWorkerThdUpdRequest struct {
	NewNoOfNbtThds ndr.DWORD
}

func (*r_WinsWorkerThdUpdRequest) Opnum() uint16 { return winsif.OpnumR_WinsWorkerThdUpd }

// r_WinsWorkerThdUpdResponse carries the [out] parameters and return value of R_WinsWorkerThdUpd.
type r_WinsWorkerThdUpdResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_WinsWorkerThdUpd calls R_WinsWorkerThdUpd (opnum 12) ([MS-RAIW] — verify the parameter
// modeling and status handling).
func R_WinsWorkerThdUpd(rpc ndr.Invoker, newNoOfNbtThds ndr.DWORD) (err error) {
	req := &r_WinsWorkerThdUpdRequest{
		NewNoOfNbtThds: newNoOfNbtThds,
	}
	var resp r_WinsWorkerThdUpdResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_WinsWorkerThdUpd: %w", err)
		return
	}
	if uint32(resp.Status) != winsif.StatusSuccess {
		err = fmt.Errorf("R_WinsWorkerThdUpd failed: %s", winsif.StatusString(uint32(resp.Status)))
	}
	return
}
