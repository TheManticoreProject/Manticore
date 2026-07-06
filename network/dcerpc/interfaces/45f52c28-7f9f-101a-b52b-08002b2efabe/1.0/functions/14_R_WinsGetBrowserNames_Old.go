package functions

// IDL source: [MS-RAIW] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-raiw/e59461f5-5486-4ec3-9ad6-14b784c1ecd6
// A fetched copy is kept at ms-raiw.idl in the interface directory.

import (
	"fmt"

	winsif "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/45f52c28-7f9f-101a-b52b-08002b2efabe/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msraiw "github.com/TheManticoreProject/Manticore/windows/protocols/ms-raiw"
)

// r_WinsGetBrowserNames_OldRequest carries the [in] parameters of R_WinsGetBrowserNames_Old.
type r_WinsGetBrowserNames_OldRequest struct {
}

func (*r_WinsGetBrowserNames_OldRequest) Opnum() uint16 { return winsif.OpnumR_WinsGetBrowserNames_Old }

// r_WinsGetBrowserNames_OldResponse carries the [out] parameters and return value of R_WinsGetBrowserNames_Old.
type r_WinsGetBrowserNames_OldResponse struct {
	PNames msraiw.WINSINTF_BROWSER_NAMES_T
	Status ndr.DWORD `ndr:"retval"`
}

// R_WinsGetBrowserNames_Old calls R_WinsGetBrowserNames_Old (opnum 14) ([MS-RAIW] — verify the parameter
// modeling and status handling).
func R_WinsGetBrowserNames_Old(rpc ndr.Invoker) (PNames msraiw.WINSINTF_BROWSER_NAMES_T, err error) {
	req := &r_WinsGetBrowserNames_OldRequest{}
	var resp r_WinsGetBrowserNames_OldResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_WinsGetBrowserNames_Old: %w", err)
		return
	}
	PNames = resp.PNames
	if uint32(resp.Status) != winsif.StatusSuccess {
		err = fmt.Errorf("R_WinsGetBrowserNames_Old failed: %s", winsif.StatusString(uint32(resp.Status)))
	}
	return
}
