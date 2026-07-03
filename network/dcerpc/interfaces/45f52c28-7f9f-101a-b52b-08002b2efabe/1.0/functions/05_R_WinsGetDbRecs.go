package functions

import (
	"fmt"

	winsif "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/45f52c28-7f9f-101a-b52b-08002b2efabe/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msraiw "github.com/TheManticoreProject/Manticore/windows/protocols/ms-raiw"
)

// r_WinsGetDbRecsRequest carries the [in] parameters of R_WinsGetDbRecs.
type r_WinsGetDbRecsRequest struct {
	PWinsAdd  msraiw.WINSINTF_ADD_T
	MinVersNo msraiw.WINSINTF_VERS_NO_T
	MaxVersNo msraiw.WINSINTF_VERS_NO_T
}

func (*r_WinsGetDbRecsRequest) Opnum() uint16 { return winsif.OpnumR_WinsGetDbRecs }

// r_WinsGetDbRecsResponse carries the [out] parameters and return value of R_WinsGetDbRecs.
type r_WinsGetDbRecsResponse struct {
	PRecs  msraiw.WINSINTF_RECS_T
	Status ndr.DWORD `ndr:"retval"`
}

// R_WinsGetDbRecs calls R_WinsGetDbRecs (opnum 5) ([MS-RAIW] — verify the parameter
// modeling and status handling).
func R_WinsGetDbRecs(rpc ndr.Invoker, pWinsAdd msraiw.WINSINTF_ADD_T, minVersNo msraiw.WINSINTF_VERS_NO_T, maxVersNo msraiw.WINSINTF_VERS_NO_T) (PRecs msraiw.WINSINTF_RECS_T, err error) {
	req := &r_WinsGetDbRecsRequest{
		PWinsAdd:  pWinsAdd,
		MinVersNo: minVersNo,
		MaxVersNo: maxVersNo,
	}
	var resp r_WinsGetDbRecsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_WinsGetDbRecs: %w", err)
		return
	}
	PRecs = resp.PRecs
	if uint32(resp.Status) != winsif.StatusSuccess {
		err = fmt.Errorf("R_WinsGetDbRecs failed: %s", winsif.StatusString(uint32(resp.Status)))
	}
	return
}
