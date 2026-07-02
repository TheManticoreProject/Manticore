package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// opnum242ReservedRequest carries the [in] parameters of Opnum242Reserved.
type opnum242ReservedRequest struct {
}

func (*opnum242ReservedRequest) Opnum() uint16 { return clusapi.OpnumOpnum242Reserved }

// opnum242ReservedResponse carries the [out] parameters and return value of Opnum242Reserved.
type opnum242ReservedResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// Opnum242Reserved calls Opnum242Reserved (opnum 242) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func Opnum242Reserved(rpc ndr.Invoker) (err error) {
	req := &opnum242ReservedRequest{}
	var resp opnum242ReservedResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("Opnum242Reserved: %w", err)
		return
	}
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("Opnum242Reserved failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
