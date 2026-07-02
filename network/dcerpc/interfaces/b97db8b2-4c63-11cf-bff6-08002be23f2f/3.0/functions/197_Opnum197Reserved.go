package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// opnum197ReservedRequest carries the [in] parameters of Opnum197Reserved.
type opnum197ReservedRequest struct {
}

func (*opnum197ReservedRequest) Opnum() uint16 { return clusapi.OpnumOpnum197Reserved }

// opnum197ReservedResponse carries the [out] parameters and return value of Opnum197Reserved.
type opnum197ReservedResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// Opnum197Reserved calls Opnum197Reserved (opnum 197) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func Opnum197Reserved(rpc ndr.Invoker) (err error) {
	req := &opnum197ReservedRequest{}
	var resp opnum197ReservedResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("Opnum197Reserved: %w", err)
		return
	}
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("Opnum197Reserved failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
