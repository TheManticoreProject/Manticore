package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// opnum189ReservedRequest carries the [in] parameters of Opnum189Reserved.
type opnum189ReservedRequest struct {
}

func (*opnum189ReservedRequest) Opnum() uint16 { return clusapi.OpnumOpnum189Reserved }

// opnum189ReservedResponse carries the [out] parameters and return value of Opnum189Reserved.
type opnum189ReservedResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// Opnum189Reserved calls Opnum189Reserved (opnum 189) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func Opnum189Reserved(rpc ndr.Invoker) (err error) {
	req := &opnum189ReservedRequest{}
	var resp opnum189ReservedResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("Opnum189Reserved: %w", err)
		return
	}
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("Opnum189Reserved failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
