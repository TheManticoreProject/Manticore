package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// opnum246ReservedRequest carries the [in] parameters of Opnum246Reserved.
type opnum246ReservedRequest struct {
}

func (*opnum246ReservedRequest) Opnum() uint16 { return clusapi.OpnumOpnum246Reserved }

// opnum246ReservedResponse carries the [out] parameters and return value of Opnum246Reserved.
type opnum246ReservedResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// Opnum246Reserved calls Opnum246Reserved (opnum 246) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func Opnum246Reserved(rpc ndr.Invoker) (err error) {
	req := &opnum246ReservedRequest{}
	var resp opnum246ReservedResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("Opnum246Reserved: %w", err)
		return
	}
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("Opnum246Reserved failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
