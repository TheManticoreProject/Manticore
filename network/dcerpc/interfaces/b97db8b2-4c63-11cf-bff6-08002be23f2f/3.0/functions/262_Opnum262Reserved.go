package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// opnum262ReservedRequest carries the [in] parameters of Opnum262Reserved.
type opnum262ReservedRequest struct {
}

func (*opnum262ReservedRequest) Opnum() uint16 { return clusapi.OpnumOpnum262Reserved }

// opnum262ReservedResponse carries the [out] parameters and return value of Opnum262Reserved.
type opnum262ReservedResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// Opnum262Reserved calls Opnum262Reserved (opnum 262) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func Opnum262Reserved(rpc ndr.Invoker) (err error) {
	req := &opnum262ReservedRequest{}
	var resp opnum262ReservedResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("Opnum262Reserved: %w", err)
		return
	}
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("Opnum262Reserved failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
