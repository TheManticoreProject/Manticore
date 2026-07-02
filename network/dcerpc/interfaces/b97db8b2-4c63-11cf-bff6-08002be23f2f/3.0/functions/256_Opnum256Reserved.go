package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// opnum256ReservedRequest carries the [in] parameters of Opnum256Reserved.
type opnum256ReservedRequest struct {
}

func (*opnum256ReservedRequest) Opnum() uint16 { return clusapi.OpnumOpnum256Reserved }

// opnum256ReservedResponse carries the [out] parameters and return value of Opnum256Reserved.
type opnum256ReservedResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// Opnum256Reserved calls Opnum256Reserved (opnum 256) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func Opnum256Reserved(rpc ndr.Invoker) (err error) {
	req := &opnum256ReservedRequest{}
	var resp opnum256ReservedResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("Opnum256Reserved: %w", err)
		return
	}
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("Opnum256Reserved failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
