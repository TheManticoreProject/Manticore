package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// opnum252ReservedRequest carries the [in] parameters of Opnum252Reserved.
type opnum252ReservedRequest struct {
}

func (*opnum252ReservedRequest) Opnum() uint16 { return clusapi.OpnumOpnum252Reserved }

// opnum252ReservedResponse carries the [out] parameters and return value of Opnum252Reserved.
type opnum252ReservedResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// Opnum252Reserved calls Opnum252Reserved (opnum 252) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func Opnum252Reserved(rpc ndr.Invoker) (err error) {
	req := &opnum252ReservedRequest{}
	var resp opnum252ReservedResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("Opnum252Reserved: %w", err)
		return
	}
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("Opnum252Reserved failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
