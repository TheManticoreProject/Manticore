package functions

import (
	"fmt"

	winsif "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/45f52c28-7f9f-101a-b52b-08002b2efabe/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msraiw "github.com/TheManticoreProject/Manticore/windows/protocols/ms-raiw"
)

// r_WinsTriggerRequest carries the [in] parameters of R_WinsTrigger.
type r_WinsTriggerRequest struct {
	PWinsAdd   msraiw.WINSINTF_ADD_T
	TrigType_e msraiw.WINSINTF_TRIG_TYPE_E
}

func (*r_WinsTriggerRequest) Opnum() uint16 { return winsif.OpnumR_WinsTrigger }

// r_WinsTriggerResponse carries the [out] parameters and return value of R_WinsTrigger.
type r_WinsTriggerResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_WinsTrigger calls R_WinsTrigger (opnum 2) ([MS-RAIW] — verify the parameter
// modeling and status handling).
func R_WinsTrigger(rpc ndr.Invoker, pWinsAdd msraiw.WINSINTF_ADD_T, trigType_e msraiw.WINSINTF_TRIG_TYPE_E) (err error) {
	req := &r_WinsTriggerRequest{
		PWinsAdd:   pWinsAdd,
		TrigType_e: trigType_e,
	}
	var resp r_WinsTriggerResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_WinsTrigger: %w", err)
		return
	}
	if uint32(resp.Status) != winsif.StatusSuccess {
		err = fmt.Errorf("R_WinsTrigger failed: %s", winsif.StatusString(uint32(resp.Status)))
	}
	return
}
