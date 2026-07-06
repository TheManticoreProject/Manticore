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

// r_WinsGetBrowserNamesRequest carries the [in] parameters of R_WinsGetBrowserNames.
//
// Unlike the other winsif methods, opnum 17 takes ServerHdl as a WINSIF_HANDLE — a
// customized ([handle]) binding handle whose base type is PWINSINTF_BIND_DATA_T. A
// customized handle serves double duty: it selects the binding AND is transmitted to
// the server as a normal [in] data parameter ([C706] Customized Handles; MIDL "handle"
// attribute). The [ref] pointer is transmitted inline, so it is modeled as an inline
// WINSINTF_BIND_DATA_T value — not omitted the way a primitive handle_t handle is.
type r_WinsGetBrowserNamesRequest struct {
	ServerHdl msraiw.WINSINTF_BIND_DATA_T
}

func (*r_WinsGetBrowserNamesRequest) Opnum() uint16 { return winsif.OpnumR_WinsGetBrowserNames }

// r_WinsGetBrowserNamesResponse carries the [out] parameters and return value of R_WinsGetBrowserNames.
type r_WinsGetBrowserNamesResponse struct {
	PNames msraiw.WINSINTF_BROWSER_NAMES_T
	Status ndr.DWORD `ndr:"retval"`
}

// R_WinsGetBrowserNames calls R_WinsGetBrowserNames (opnum 17) ([MS-RAIW] 3.1.4.18).
func R_WinsGetBrowserNames(rpc ndr.Invoker, serverHdl msraiw.WINSINTF_BIND_DATA_T) (PNames msraiw.WINSINTF_BROWSER_NAMES_T, err error) {
	req := &r_WinsGetBrowserNamesRequest{
		ServerHdl: serverHdl,
	}
	var resp r_WinsGetBrowserNamesResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_WinsGetBrowserNames: %w", err)
		return
	}
	PNames = resp.PNames
	if uint32(resp.Status) != winsif.StatusSuccess {
		err = fmt.Errorf("R_WinsGetBrowserNames failed: %s", winsif.StatusString(uint32(resp.Status)))
	}
	return
}
