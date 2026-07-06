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

// r_WinsStatusRequest carries the [in] parameters of R_WinsStatus.
type r_WinsStatusRequest struct {
	Cmd_e    msraiw.WINSINTF_CMD_E
	PResults msraiw.WINSINTF_RESULTS_T
}

func (*r_WinsStatusRequest) Opnum() uint16 { return winsif.OpnumR_WinsStatus }

// r_WinsStatusResponse carries the [out] parameters and return value of R_WinsStatus.
type r_WinsStatusResponse struct {
	PResults msraiw.WINSINTF_RESULTS_T
	Status   ndr.DWORD `ndr:"retval"`
}

// R_WinsStatus calls R_WinsStatus (opnum 1) ([MS-RAIW] — verify the parameter
// modeling and status handling).
func R_WinsStatus(rpc ndr.Invoker, cmd_e msraiw.WINSINTF_CMD_E, pResults msraiw.WINSINTF_RESULTS_T) (PResults msraiw.WINSINTF_RESULTS_T, err error) {
	req := &r_WinsStatusRequest{
		Cmd_e:    cmd_e,
		PResults: pResults,
	}
	var resp r_WinsStatusResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_WinsStatus: %w", err)
		return
	}
	PResults = resp.PResults
	if uint32(resp.Status) != winsif.StatusSuccess {
		err = fmt.Errorf("R_WinsStatus failed: %s", winsif.StatusString(uint32(resp.Status)))
	}
	return
}
