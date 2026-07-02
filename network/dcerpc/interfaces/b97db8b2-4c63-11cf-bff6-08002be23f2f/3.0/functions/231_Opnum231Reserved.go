package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// opnum231ReservedRequest carries the [in] parameters of Opnum231Reserved.
type opnum231ReservedRequest struct {
}

func (*opnum231ReservedRequest) Opnum() uint16 { return clusapi.OpnumOpnum231Reserved }

// opnum231ReservedResponse carries the [out] parameters and return value of Opnum231Reserved.
type opnum231ReservedResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// Opnum231Reserved calls Opnum231Reserved (opnum 231) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func Opnum231Reserved(rpc ndr.Invoker) (err error) {
	req := &opnum231ReservedRequest{}
	var resp opnum231ReservedResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("Opnum231Reserved: %w", err)
		return
	}
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("Opnum231Reserved failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
