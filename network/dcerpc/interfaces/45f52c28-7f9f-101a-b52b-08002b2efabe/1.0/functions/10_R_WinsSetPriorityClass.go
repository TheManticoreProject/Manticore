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

// r_WinsSetPriorityClassRequest carries the [in] parameters of R_WinsSetPriorityClass.
type r_WinsSetPriorityClassRequest struct {
	PrCls_e msraiw.WINSINTF_PRIORITY_CLASS_E
}

func (*r_WinsSetPriorityClassRequest) Opnum() uint16 { return winsif.OpnumR_WinsSetPriorityClass }

// r_WinsSetPriorityClassResponse carries the [out] parameters and return value of R_WinsSetPriorityClass.
type r_WinsSetPriorityClassResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// R_WinsSetPriorityClass calls R_WinsSetPriorityClass (opnum 10) ([MS-RAIW] — verify the parameter
// modeling and status handling).
func R_WinsSetPriorityClass(rpc ndr.Invoker, prCls_e msraiw.WINSINTF_PRIORITY_CLASS_E) (err error) {
	req := &r_WinsSetPriorityClassRequest{
		PrCls_e: prCls_e,
	}
	var resp r_WinsSetPriorityClassResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_WinsSetPriorityClass: %w", err)
		return
	}
	if uint32(resp.Status) != winsif.StatusSuccess {
		err = fmt.Errorf("R_WinsSetPriorityClass failed: %s", winsif.StatusString(uint32(resp.Status)))
	}
	return
}
