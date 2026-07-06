package functions

// IDL source: [MS-CMRP] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cmrp/e12b6d8f-c410-49d7-a27d-9992782a9027
// A fetched copy is kept at ms-cmrp.idl in the interface directory.

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// opnum248ReservedRequest carries the [in] parameters of Opnum248Reserved.
type opnum248ReservedRequest struct {
}

func (*opnum248ReservedRequest) Opnum() uint16 { return clusapi.Opnum248Reserved }

// opnum248ReservedResponse carries the [out] parameters and return value of Opnum248Reserved.
type opnum248ReservedResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// Opnum248Reserved calls Opnum248Reserved (opnum 248) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func Opnum248Reserved(rpc ndr.Invoker) (err error) {
	req := &opnum248ReservedRequest{}
	var resp opnum248ReservedResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("Opnum248Reserved: %w", err)
		return
	}
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("Opnum248Reserved failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
