package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// opnum263ReservedRequest carries the [in] parameters of Opnum263Reserved.
type opnum263ReservedRequest struct {
}

func (*opnum263ReservedRequest) Opnum() uint16 { return clusapi.OpnumOpnum263Reserved }

// opnum263ReservedResponse carries the [out] parameters and return value of Opnum263Reserved.
type opnum263ReservedResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// Opnum263Reserved calls Opnum263Reserved (opnum 263) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func Opnum263Reserved(rpc ndr.Invoker) (err error) {
	req := &opnum263ReservedRequest{}
	var resp opnum263ReservedResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("Opnum263Reserved: %w", err)
		return
	}
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("Opnum263Reserved failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
