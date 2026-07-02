package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// opnum244ReservedRequest carries the [in] parameters of Opnum244Reserved.
type opnum244ReservedRequest struct {
}

func (*opnum244ReservedRequest) Opnum() uint16 { return clusapi.OpnumOpnum244Reserved }

// opnum244ReservedResponse carries the [out] parameters and return value of Opnum244Reserved.
type opnum244ReservedResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// Opnum244Reserved calls Opnum244Reserved (opnum 244) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func Opnum244Reserved(rpc ndr.Invoker) (err error) {
	req := &opnum244ReservedRequest{}
	var resp opnum244ReservedResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("Opnum244Reserved: %w", err)
		return
	}
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("Opnum244Reserved failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
