package functions

import (
	"fmt"

	winsif "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/45f52c28-7f9f-101a-b52b-08002b2efabe/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msraiw "github.com/TheManticoreProject/Manticore/windows/protocols/ms-raiw"
)

// r_WinsStatusWHdlRequest carries the [in] parameters of R_WinsStatusWHdl.
//
// ServerHdl is a WINSIF_HANDLE — a customized ([handle]) binding handle over
// PWINSINTF_BIND_DATA_T. A customized handle both selects the binding and is
// transmitted to the server as the first [in] data parameter ([C706] Customized
// Handles; MIDL "handle" attribute), so it precedes Cmd_e on the wire as an inline
// WINSINTF_BIND_DATA_T value.
type r_WinsStatusWHdlRequest struct {
	ServerHdl msraiw.WINSINTF_BIND_DATA_T
	Cmd_e     msraiw.WINSINTF_CMD_E
	PResults  msraiw.WINSINTF_RESULTS_NEW_T
}

func (*r_WinsStatusWHdlRequest) Opnum() uint16 { return winsif.OpnumR_WinsStatusWHdl }

// r_WinsStatusWHdlResponse carries the [out] parameters and return value of R_WinsStatusWHdl.
type r_WinsStatusWHdlResponse struct {
	PResults msraiw.WINSINTF_RESULTS_NEW_T
	Status   ndr.DWORD `ndr:"retval"`
}

// R_WinsStatusWHdl calls R_WinsStatusWHdl (opnum 20) ([MS-RAIW] 3.1.4.21).
func R_WinsStatusWHdl(rpc ndr.Invoker, serverHdl msraiw.WINSINTF_BIND_DATA_T, cmd_e msraiw.WINSINTF_CMD_E, pResults msraiw.WINSINTF_RESULTS_NEW_T) (PResults msraiw.WINSINTF_RESULTS_NEW_T, err error) {
	req := &r_WinsStatusWHdlRequest{
		ServerHdl: serverHdl,
		Cmd_e:     cmd_e,
		PResults:  pResults,
	}
	var resp r_WinsStatusWHdlResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_WinsStatusWHdl: %w", err)
		return
	}
	PResults = resp.PResults
	if uint32(resp.Status) != winsif.StatusSuccess {
		err = fmt.Errorf("R_WinsStatusWHdl failed: %s", winsif.StatusString(uint32(resp.Status)))
	}
	return
}
