package functions

import (
	"fmt"

	winsif "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/45f52c28-7f9f-101a-b52b-08002b2efabe/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msraiw "github.com/TheManticoreProject/Manticore/windows/protocols/ms-raiw"
)

// r_WinsDelDbRecsRequest carries the [in] parameters of R_WinsDelDbRecs.
type r_WinsDelDbRecsRequest struct {
	PWinsAdd  msraiw.WINSINTF_ADD_T
	MinVersNo msraiw.WINSINTF_VERS_NO_T
	MaxVersNo msraiw.WINSINTF_VERS_NO_T
}

func (*r_WinsDelDbRecsRequest) Opnum() uint16 { return winsif.OpnumR_WinsDelDbRecs }

// r_WinsDelDbRecsResponse carries the [out] parameters and return value of R_WinsDelDbRecs.
type r_WinsDelDbRecsResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_WinsDelDbRecs calls R_WinsDelDbRecs (opnum 8) ([MS-RAIW] — verify the parameter
// modeling and status handling).
func R_WinsDelDbRecs(rpc ndr.Invoker, pWinsAdd msraiw.WINSINTF_ADD_T, minVersNo msraiw.WINSINTF_VERS_NO_T, maxVersNo msraiw.WINSINTF_VERS_NO_T) (err error) {
	req := &r_WinsDelDbRecsRequest{
		PWinsAdd:  pWinsAdd,
		MinVersNo: minVersNo,
		MaxVersNo: maxVersNo,
	}
	var resp r_WinsDelDbRecsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_WinsDelDbRecs: %w", err)
		return
	}
	if uint32(resp.Status) != winsif.StatusSuccess {
		err = fmt.Errorf("R_WinsDelDbRecs failed: %s", winsif.StatusString(uint32(resp.Status)))
	}
	return
}
