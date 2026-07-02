package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// opnum192ReservedRequest carries the [in] parameters of Opnum192Reserved.
type opnum192ReservedRequest struct {
}

func (*opnum192ReservedRequest) Opnum() uint16 { return clusapi.OpnumOpnum192Reserved }

// opnum192ReservedResponse carries the [out] parameters and return value of Opnum192Reserved.
type opnum192ReservedResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// Opnum192Reserved calls Opnum192Reserved (opnum 192) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func Opnum192Reserved(rpc ndr.Invoker) (err error) {
	req := &opnum192ReservedRequest{}
	var resp opnum192ReservedResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("Opnum192Reserved: %w", err)
		return
	}
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("Opnum192Reserved failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
