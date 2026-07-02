package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// opnum217ReservedRequest carries the [in] parameters of Opnum217Reserved.
type opnum217ReservedRequest struct {
}

func (*opnum217ReservedRequest) Opnum() uint16 { return clusapi.OpnumOpnum217Reserved }

// opnum217ReservedResponse carries the [out] parameters and return value of Opnum217Reserved.
type opnum217ReservedResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// Opnum217Reserved calls Opnum217Reserved (opnum 217) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func Opnum217Reserved(rpc ndr.Invoker) (err error) {
	req := &opnum217ReservedRequest{}
	var resp opnum217ReservedResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("Opnum217Reserved: %w", err)
		return
	}
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("Opnum217Reserved failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
