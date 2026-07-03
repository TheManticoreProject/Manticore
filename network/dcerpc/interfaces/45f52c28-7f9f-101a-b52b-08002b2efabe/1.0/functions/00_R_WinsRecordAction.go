package functions

import (
	"fmt"

	winsif "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/45f52c28-7f9f-101a-b52b-08002b2efabe/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msraiw "github.com/TheManticoreProject/Manticore/windows/protocols/ms-raiw"
)

// r_WinsRecordActionRequest carries the [in] parameters of R_WinsRecordAction.
type r_WinsRecordActionRequest struct {
	PpRecAction *msraiw.WINSINTF_RECORD_ACTION_T `ndr:"unique"`
}

func (*r_WinsRecordActionRequest) Opnum() uint16 { return winsif.OpnumR_WinsRecordAction }

// r_WinsRecordActionResponse carries the [out] parameters and return value of R_WinsRecordAction.
type r_WinsRecordActionResponse struct {
	PpRecAction *msraiw.WINSINTF_RECORD_ACTION_T `ndr:"unique"`
	Status      ndr.DWORD                        `ndr:"retval"`
}

// R_WinsRecordAction calls R_WinsRecordAction (opnum 0) ([MS-RAIW] — verify the parameter
// modeling and status handling).
func R_WinsRecordAction(rpc ndr.Invoker, ppRecAction *msraiw.WINSINTF_RECORD_ACTION_T) (PpRecAction *msraiw.WINSINTF_RECORD_ACTION_T, err error) {
	req := &r_WinsRecordActionRequest{
		PpRecAction: ppRecAction,
	}
	var resp r_WinsRecordActionResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_WinsRecordAction: %w", err)
		return
	}
	PpRecAction = resp.PpRecAction
	if uint32(resp.Status) != winsif.StatusSuccess {
		err = fmt.Errorf("R_WinsRecordAction failed: %s", winsif.StatusString(uint32(resp.Status)))
	}
	return
}
