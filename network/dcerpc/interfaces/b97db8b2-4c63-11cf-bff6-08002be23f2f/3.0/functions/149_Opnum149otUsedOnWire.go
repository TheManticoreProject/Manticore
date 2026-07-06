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

// opnum149otUsedOnWireRequest carries the [in] parameters of Opnum149otUsedOnWire.
type opnum149otUsedOnWireRequest struct {
}

func (*opnum149otUsedOnWireRequest) Opnum() uint16 { return clusapi.OpnumOpnum149otUsedOnWire }

// opnum149otUsedOnWireResponse carries the [out] parameters and return value of Opnum149otUsedOnWire.
type opnum149otUsedOnWireResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// Opnum149otUsedOnWire calls Opnum149otUsedOnWire (opnum 149) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func Opnum149otUsedOnWire(rpc ndr.Invoker) (err error) {
	req := &opnum149otUsedOnWireRequest{}
	var resp opnum149otUsedOnWireResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("Opnum149otUsedOnWire: %w", err)
		return
	}
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("Opnum149otUsedOnWire failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
