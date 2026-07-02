package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// opnum261ReservedRequest carries the [in] parameters of Opnum261Reserved.
type opnum261ReservedRequest struct {
}

func (*opnum261ReservedRequest) Opnum() uint16 { return clusapi.OpnumOpnum261Reserved }

// opnum261ReservedResponse carries the [out] parameters and return value of Opnum261Reserved.
type opnum261ReservedResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// Opnum261Reserved calls Opnum261Reserved (opnum 261) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func Opnum261Reserved(rpc ndr.Invoker) (err error) {
	req := &opnum261ReservedRequest{}
	var resp opnum261ReservedResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("Opnum261Reserved: %w", err)
		return
	}
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("Opnum261Reserved failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
