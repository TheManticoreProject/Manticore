package functions

import (
	"fmt"

	winsif "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/45f52c28-7f9f-101a-b52b-08002b2efabe/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msraiw "github.com/TheManticoreProject/Manticore/windows/protocols/ms-raiw"
)

// r_WinsPullRangeRequest carries the [in] parameters of R_WinsPullRange.
type r_WinsPullRangeRequest struct {
	PWinsAdd  msraiw.WINSINTF_ADD_T
	POwnerAdd msraiw.WINSINTF_ADD_T
	MinVersNo msraiw.WINSINTF_VERS_NO_T
	MaxVersNo msraiw.WINSINTF_VERS_NO_T
}

func (*r_WinsPullRangeRequest) Opnum() uint16 { return winsif.OpnumR_WinsPullRange }

// r_WinsPullRangeResponse carries the [out] parameters and return value of R_WinsPullRange.
type r_WinsPullRangeResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_WinsPullRange calls R_WinsPullRange (opnum 9) ([MS-RAIW] — verify the parameter
// modeling and status handling).
func R_WinsPullRange(rpc ndr.Invoker, pWinsAdd msraiw.WINSINTF_ADD_T, pOwnerAdd msraiw.WINSINTF_ADD_T, minVersNo msraiw.WINSINTF_VERS_NO_T, maxVersNo msraiw.WINSINTF_VERS_NO_T) (err error) {
	req := &r_WinsPullRangeRequest{
		PWinsAdd:  pWinsAdd,
		POwnerAdd: pOwnerAdd,
		MinVersNo: minVersNo,
		MaxVersNo: maxVersNo,
	}
	var resp r_WinsPullRangeResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_WinsPullRange: %w", err)
		return
	}
	if uint32(resp.Status) != winsif.StatusSuccess {
		err = fmt.Errorf("R_WinsPullRange failed: %s", winsif.StatusString(uint32(resp.Status)))
	}
	return
}
