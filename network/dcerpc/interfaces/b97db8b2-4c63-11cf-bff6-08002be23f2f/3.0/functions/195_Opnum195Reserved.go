package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// opnum195ReservedRequest carries the [in] parameters of Opnum195Reserved.
type opnum195ReservedRequest struct {
}

func (*opnum195ReservedRequest) Opnum() uint16 { return clusapi.OpnumOpnum195Reserved }

// opnum195ReservedResponse carries the [out] parameters and return value of Opnum195Reserved.
type opnum195ReservedResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// Opnum195Reserved calls Opnum195Reserved (opnum 195) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func Opnum195Reserved(rpc ndr.Invoker) (err error) {
	req := &opnum195ReservedRequest{}
	var resp opnum195ReservedResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("Opnum195Reserved: %w", err)
		return
	}
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("Opnum195Reserved failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
