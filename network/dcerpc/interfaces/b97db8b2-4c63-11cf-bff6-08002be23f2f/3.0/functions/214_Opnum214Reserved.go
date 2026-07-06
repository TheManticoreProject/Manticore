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

// opnum214ReservedRequest carries the [in] parameters of Opnum214Reserved.
type opnum214ReservedRequest struct {
}

func (*opnum214ReservedRequest) Opnum() uint16 { return clusapi.Opnum214Reserved }

// opnum214ReservedResponse carries the [out] parameters and return value of Opnum214Reserved.
type opnum214ReservedResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// Opnum214Reserved calls Opnum214Reserved (opnum 214) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func Opnum214Reserved(rpc ndr.Invoker) (err error) {
	req := &opnum214ReservedRequest{}
	var resp opnum214ReservedResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("Opnum214Reserved: %w", err)
		return
	}
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("Opnum214Reserved failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
