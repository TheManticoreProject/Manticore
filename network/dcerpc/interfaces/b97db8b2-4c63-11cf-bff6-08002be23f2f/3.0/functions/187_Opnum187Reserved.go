package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// opnum187ReservedRequest carries the [in] parameters of Opnum187Reserved.
type opnum187ReservedRequest struct {
}

func (*opnum187ReservedRequest) Opnum() uint16 { return clusapi.OpnumOpnum187Reserved }

// opnum187ReservedResponse carries the [out] parameters and return value of Opnum187Reserved.
type opnum187ReservedResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// Opnum187Reserved calls Opnum187Reserved (opnum 187) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func Opnum187Reserved(rpc ndr.Invoker) (err error) {
	req := &opnum187ReservedRequest{}
	var resp opnum187ReservedResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("Opnum187Reserved: %w", err)
		return
	}
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("Opnum187Reserved failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
