package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// opnum220ReservedRequest carries the [in] parameters of Opnum220Reserved.
type opnum220ReservedRequest struct {
}

func (*opnum220ReservedRequest) Opnum() uint16 { return clusapi.OpnumOpnum220Reserved }

// opnum220ReservedResponse carries the [out] parameters and return value of Opnum220Reserved.
type opnum220ReservedResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// Opnum220Reserved calls Opnum220Reserved (opnum 220) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func Opnum220Reserved(rpc ndr.Invoker) (err error) {
	req := &opnum220ReservedRequest{}
	var resp opnum220ReservedResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("Opnum220Reserved: %w", err)
		return
	}
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("Opnum220Reserved failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
