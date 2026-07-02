package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// opnum233ReservedRequest carries the [in] parameters of Opnum233Reserved.
type opnum233ReservedRequest struct {
}

func (*opnum233ReservedRequest) Opnum() uint16 { return clusapi.OpnumOpnum233Reserved }

// opnum233ReservedResponse carries the [out] parameters and return value of Opnum233Reserved.
type opnum233ReservedResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// Opnum233Reserved calls Opnum233Reserved (opnum 233) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func Opnum233Reserved(rpc ndr.Invoker) (err error) {
	req := &opnum233ReservedRequest{}
	var resp opnum233ReservedResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("Opnum233Reserved: %w", err)
		return
	}
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("Opnum233Reserved failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
