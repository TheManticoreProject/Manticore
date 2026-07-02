package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// opnum215ReservedRequest carries the [in] parameters of Opnum215Reserved.
type opnum215ReservedRequest struct {
}

func (*opnum215ReservedRequest) Opnum() uint16 { return clusapi.OpnumOpnum215Reserved }

// opnum215ReservedResponse carries the [out] parameters and return value of Opnum215Reserved.
type opnum215ReservedResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// Opnum215Reserved calls Opnum215Reserved (opnum 215) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func Opnum215Reserved(rpc ndr.Invoker) (err error) {
	req := &opnum215ReservedRequest{}
	var resp opnum215ReservedResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("Opnum215Reserved: %w", err)
		return
	}
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("Opnum215Reserved failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
