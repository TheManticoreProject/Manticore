package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// opnum226ReservedRequest carries the [in] parameters of Opnum226Reserved.
type opnum226ReservedRequest struct {
}

func (*opnum226ReservedRequest) Opnum() uint16 { return clusapi.OpnumOpnum226Reserved }

// opnum226ReservedResponse carries the [out] parameters and return value of Opnum226Reserved.
type opnum226ReservedResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// Opnum226Reserved calls Opnum226Reserved (opnum 226) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func Opnum226Reserved(rpc ndr.Invoker) (err error) {
	req := &opnum226ReservedRequest{}
	var resp opnum226ReservedResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("Opnum226Reserved: %w", err)
		return
	}
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("Opnum226Reserved failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
