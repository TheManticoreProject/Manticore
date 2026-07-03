package functions

import (
	"fmt"

	winsif "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/45f52c28-7f9f-101a-b52b-08002b2efabe/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// r_WinsSetFlagsRequest carries the [in] parameters of R_WinsSetFlags.
type r_WinsSetFlagsRequest struct {
	FFlags ndr.DWORD
}

func (*r_WinsSetFlagsRequest) Opnum() uint16 { return winsif.OpnumR_WinsSetFlags }

// r_WinsSetFlagsResponse carries the [out] parameters and return value of R_WinsSetFlags.
type r_WinsSetFlagsResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_WinsSetFlags calls R_WinsSetFlags (opnum 16) ([MS-RAIW] — verify the parameter
// modeling and status handling).
func R_WinsSetFlags(rpc ndr.Invoker, fFlags ndr.DWORD) (err error) {
	req := &r_WinsSetFlagsRequest{
		FFlags: fFlags,
	}
	var resp r_WinsSetFlagsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_WinsSetFlags: %w", err)
		return
	}
	if uint32(resp.Status) != winsif.StatusSuccess {
		err = fmt.Errorf("R_WinsSetFlags failed: %s", winsif.StatusString(uint32(resp.Status)))
	}
	return
}
