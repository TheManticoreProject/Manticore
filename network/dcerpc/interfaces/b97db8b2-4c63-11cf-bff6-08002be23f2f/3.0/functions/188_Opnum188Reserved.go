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

// opnum188ReservedRequest carries the [in] parameters of Opnum188Reserved.
type opnum188ReservedRequest struct {
}

func (*opnum188ReservedRequest) Opnum() uint16 { return clusapi.OpnumOpnum188Reserved }

// opnum188ReservedResponse carries the [out] parameters and return value of Opnum188Reserved.
type opnum188ReservedResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// Opnum188Reserved calls Opnum188Reserved (opnum 188) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func Opnum188Reserved(rpc ndr.Invoker) (err error) {
	req := &opnum188ReservedRequest{}
	var resp opnum188ReservedResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("Opnum188Reserved: %w", err)
		return
	}
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("Opnum188Reserved failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
