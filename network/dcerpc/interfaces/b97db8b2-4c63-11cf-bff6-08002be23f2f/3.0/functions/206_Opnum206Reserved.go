package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// opnum206ReservedRequest carries the [in] parameters of Opnum206Reserved.
type opnum206ReservedRequest struct {
}

func (*opnum206ReservedRequest) Opnum() uint16 { return clusapi.OpnumOpnum206Reserved }

// opnum206ReservedResponse carries the [out] parameters and return value of Opnum206Reserved.
type opnum206ReservedResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// Opnum206Reserved calls Opnum206Reserved (opnum 206) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func Opnum206Reserved(rpc ndr.Invoker) (err error) {
	req := &opnum206ReservedRequest{}
	var resp opnum206ReservedResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("Opnum206Reserved: %w", err)
		return
	}
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("Opnum206Reserved failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
