package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// opnum254ReservedRequest carries the [in] parameters of Opnum254Reserved.
type opnum254ReservedRequest struct {
}

func (*opnum254ReservedRequest) Opnum() uint16 { return clusapi.OpnumOpnum254Reserved }

// opnum254ReservedResponse carries the [out] parameters and return value of Opnum254Reserved.
type opnum254ReservedResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// Opnum254Reserved calls Opnum254Reserved (opnum 254) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func Opnum254Reserved(rpc ndr.Invoker) (err error) {
	req := &opnum254ReservedRequest{}
	var resp opnum254ReservedResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("Opnum254Reserved: %w", err)
		return
	}
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("Opnum254Reserved failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
