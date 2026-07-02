package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// opnum194ReservedRequest carries the [in] parameters of Opnum194Reserved.
type opnum194ReservedRequest struct {
}

func (*opnum194ReservedRequest) Opnum() uint16 { return clusapi.OpnumOpnum194Reserved }

// opnum194ReservedResponse carries the [out] parameters and return value of Opnum194Reserved.
type opnum194ReservedResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// Opnum194Reserved calls Opnum194Reserved (opnum 194) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func Opnum194Reserved(rpc ndr.Invoker) (err error) {
	req := &opnum194ReservedRequest{}
	var resp opnum194ReservedResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("Opnum194Reserved: %w", err)
		return
	}
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("Opnum194Reserved failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
