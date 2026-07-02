package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// opnum239ReservedRequest carries the [in] parameters of Opnum239Reserved.
type opnum239ReservedRequest struct {
}

func (*opnum239ReservedRequest) Opnum() uint16 { return clusapi.OpnumOpnum239Reserved }

// opnum239ReservedResponse carries the [out] parameters and return value of Opnum239Reserved.
type opnum239ReservedResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// Opnum239Reserved calls Opnum239Reserved (opnum 239) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func Opnum239Reserved(rpc ndr.Invoker) (err error) {
	req := &opnum239ReservedRequest{}
	var resp opnum239ReservedResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("Opnum239Reserved: %w", err)
		return
	}
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("Opnum239Reserved failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
