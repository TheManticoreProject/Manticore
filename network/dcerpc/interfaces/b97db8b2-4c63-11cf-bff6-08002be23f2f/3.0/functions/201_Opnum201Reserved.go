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

// opnum201ReservedRequest carries the [in] parameters of Opnum201Reserved.
type opnum201ReservedRequest struct {
}

func (*opnum201ReservedRequest) Opnum() uint16 { return clusapi.Opnum201Reserved }

// opnum201ReservedResponse carries the [out] parameters and return value of Opnum201Reserved.
type opnum201ReservedResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// Opnum201Reserved calls Opnum201Reserved (opnum 201) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func Opnum201Reserved(rpc ndr.Invoker) (err error) {
	req := &opnum201ReservedRequest{}
	var resp opnum201ReservedResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("Opnum201Reserved: %w", err)
		return
	}
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("Opnum201Reserved failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
