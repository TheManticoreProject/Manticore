package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// opnum236ReservedRequest carries the [in] parameters of Opnum236Reserved.
type opnum236ReservedRequest struct {
}

func (*opnum236ReservedRequest) Opnum() uint16 { return clusapi.OpnumOpnum236Reserved }

// opnum236ReservedResponse carries the [out] parameters and return value of Opnum236Reserved.
type opnum236ReservedResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// Opnum236Reserved calls Opnum236Reserved (opnum 236) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func Opnum236Reserved(rpc ndr.Invoker) (err error) {
	req := &opnum236ReservedRequest{}
	var resp opnum236ReservedResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("Opnum236Reserved: %w", err)
		return
	}
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("Opnum236Reserved failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
