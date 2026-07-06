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

// opnum222ReservedRequest carries the [in] parameters of Opnum222Reserved.
type opnum222ReservedRequest struct {
}

func (*opnum222ReservedRequest) Opnum() uint16 { return clusapi.Opnum222Reserved }

// opnum222ReservedResponse carries the [out] parameters and return value of Opnum222Reserved.
type opnum222ReservedResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// Opnum222Reserved calls Opnum222Reserved (opnum 222) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func Opnum222Reserved(rpc ndr.Invoker) (err error) {
	req := &opnum222ReservedRequest{}
	var resp opnum222ReservedResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("Opnum222Reserved: %w", err)
		return
	}
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("Opnum222Reserved failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
