package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// opnum204ReservedRequest carries the [in] parameters of Opnum204Reserved.
type opnum204ReservedRequest struct {
}

func (*opnum204ReservedRequest) Opnum() uint16 { return clusapi.OpnumOpnum204Reserved }

// opnum204ReservedResponse carries the [out] parameters and return value of Opnum204Reserved.
type opnum204ReservedResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// Opnum204Reserved calls Opnum204Reserved (opnum 204) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func Opnum204Reserved(rpc ndr.Invoker) (err error) {
	req := &opnum204ReservedRequest{}
	var resp opnum204ReservedResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("Opnum204Reserved: %w", err)
		return
	}
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("Opnum204Reserved failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
