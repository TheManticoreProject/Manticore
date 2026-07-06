package functions

// IDL source: [MS-RAIW] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-raiw/e59461f5-5486-4ec3-9ad6-14b784c1ecd6
// A fetched copy is kept at ms-raiw.idl in the interface directory.

import (
	"fmt"

	winsi2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/811109bf-a4e1-11d1-ab54-00a0c91e9b45/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msraiw "github.com/TheManticoreProject/Manticore/windows/protocols/ms-raiw"
)

// r_WinsTombstoneDbRecsRequest carries the [in] parameters of R_WinsTombstoneDbRecs.
type r_WinsTombstoneDbRecsRequest struct {
	PWinsAdd  msraiw.WINSINTF_ADD_T
	MinVersNo msraiw.WINSINTF_VERS_NO_T
	MaxVersNo msraiw.WINSINTF_VERS_NO_T
}

func (*r_WinsTombstoneDbRecsRequest) Opnum() uint16 { return winsi2.OpnumR_WinsTombstoneDbRecs }

// r_WinsTombstoneDbRecsResponse carries the [out] parameters and return value of R_WinsTombstoneDbRecs.
type r_WinsTombstoneDbRecsResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_WinsTombstoneDbRecs calls R_WinsTombstoneDbRecs (opnum 0) ([MS-RAIW] 3.2.4.1).
// pWinsAdd is a [ref] pointer, transmitted inline; the version numbers are passed by
// value.
func R_WinsTombstoneDbRecs(rpc ndr.Invoker, pWinsAdd msraiw.WINSINTF_ADD_T, minVersNo msraiw.WINSINTF_VERS_NO_T, maxVersNo msraiw.WINSINTF_VERS_NO_T) (err error) {
	req := &r_WinsTombstoneDbRecsRequest{
		PWinsAdd:  pWinsAdd,
		MinVersNo: minVersNo,
		MaxVersNo: maxVersNo,
	}
	var resp r_WinsTombstoneDbRecsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_WinsTombstoneDbRecs: %w", err)
		return
	}
	if uint32(resp.Status) != winsi2.StatusSuccess {
		err = fmt.Errorf("R_WinsTombstoneDbRecs failed: %s", winsi2.StatusString(uint32(resp.Status)))
	}
	return
}
