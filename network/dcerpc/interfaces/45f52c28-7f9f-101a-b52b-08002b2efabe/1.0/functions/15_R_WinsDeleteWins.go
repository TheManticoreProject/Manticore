package functions

import (
	"fmt"

	winsif "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/45f52c28-7f9f-101a-b52b-08002b2efabe/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msraiw "github.com/TheManticoreProject/Manticore/windows/protocols/ms-raiw"
)

// r_WinsDeleteWinsRequest carries the [in] parameters of R_WinsDeleteWins.
type r_WinsDeleteWinsRequest struct {
	PWinsAdd msraiw.WINSINTF_ADD_T
}

func (*r_WinsDeleteWinsRequest) Opnum() uint16 { return winsif.OpnumR_WinsDeleteWins }

// r_WinsDeleteWinsResponse carries the [out] parameters and return value of R_WinsDeleteWins.
type r_WinsDeleteWinsResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_WinsDeleteWins calls R_WinsDeleteWins (opnum 15) ([MS-RAIW] — verify the parameter
// modeling and status handling).
func R_WinsDeleteWins(rpc ndr.Invoker, pWinsAdd msraiw.WINSINTF_ADD_T) (err error) {
	req := &r_WinsDeleteWinsRequest{
		PWinsAdd: pWinsAdd,
	}
	var resp r_WinsDeleteWinsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_WinsDeleteWins: %w", err)
		return
	}
	if uint32(resp.Status) != winsif.StatusSuccess {
		err = fmt.Errorf("R_WinsDeleteWins failed: %s", winsif.StatusString(uint32(resp.Status)))
	}
	return
}
