package functions

import (
	"fmt"

	winsif "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/45f52c28-7f9f-101a-b52b-08002b2efabe/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msraiw "github.com/TheManticoreProject/Manticore/windows/protocols/ms-raiw"
)

// r_WinsStatusNewRequest carries the [in] parameters of R_WinsStatusNew.
type r_WinsStatusNewRequest struct {
	Cmd_e msraiw.WINSINTF_CMD_E
}

func (*r_WinsStatusNewRequest) Opnum() uint16 { return winsif.OpnumR_WinsStatusNew }

// r_WinsStatusNewResponse carries the [out] parameters and return value of R_WinsStatusNew.
type r_WinsStatusNewResponse struct {
	PResults msraiw.WINSINTF_RESULTS_NEW_T
	Status   ndr.DWORD `ndr:"retval"`
}

// R_WinsStatusNew calls R_WinsStatusNew (opnum 19) ([MS-RAIW] — verify the parameter
// modeling and status handling).
func R_WinsStatusNew(rpc ndr.Invoker, cmd_e msraiw.WINSINTF_CMD_E) (PResults msraiw.WINSINTF_RESULTS_NEW_T, err error) {
	req := &r_WinsStatusNewRequest{
		Cmd_e: cmd_e,
	}
	var resp r_WinsStatusNewResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_WinsStatusNew: %w", err)
		return
	}
	PResults = resp.PResults
	if uint32(resp.Status) != winsif.StatusSuccess {
		err = fmt.Errorf("R_WinsStatusNew failed: %s", winsif.StatusString(uint32(resp.Status)))
	}
	return
}
