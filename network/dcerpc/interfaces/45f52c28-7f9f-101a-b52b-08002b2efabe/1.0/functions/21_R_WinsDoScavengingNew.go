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

// r_WinsDoScavengingNewRequest carries the [in] parameters of R_WinsDoScavengingNew.
type r_WinsDoScavengingNewRequest struct {
	PScvReq msraiw.WINSINTF_SCV_REQ_T
}

func (*r_WinsDoScavengingNewRequest) Opnum() uint16 { return winsif.OpnumR_WinsDoScavengingNew }

// r_WinsDoScavengingNewResponse carries the [out] parameters and return value of R_WinsDoScavengingNew.
type r_WinsDoScavengingNewResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_WinsDoScavengingNew calls R_WinsDoScavengingNew (opnum 21) ([MS-RAIW] — verify the parameter
// modeling and status handling).
func R_WinsDoScavengingNew(rpc ndr.Invoker, pScvReq msraiw.WINSINTF_SCV_REQ_T) (err error) {
	req := &r_WinsDoScavengingNewRequest{
		PScvReq: pScvReq,
	}
	var resp r_WinsDoScavengingNewResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_WinsDoScavengingNew: %w", err)
		return
	}
	if uint32(resp.Status) != winsif.StatusSuccess {
		err = fmt.Errorf("R_WinsDoScavengingNew failed: %s", winsif.StatusString(uint32(resp.Status)))
	}
	return
}
