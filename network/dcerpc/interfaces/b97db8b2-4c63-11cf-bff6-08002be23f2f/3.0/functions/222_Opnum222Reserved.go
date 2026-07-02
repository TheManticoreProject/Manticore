package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// opnum222ReservedRequest carries the [in] parameters of Opnum222Reserved.
type opnum222ReservedRequest struct {
}

func (*opnum222ReservedRequest) Opnum() uint16 { return clusapi.OpnumOpnum222Reserved }

// opnum222ReservedResponse carries the [out] parameters and return value of Opnum222Reserved.
type opnum222ReservedResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// Opnum222Reserved calls Opnum222Reserved (opnum 222) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func Opnum222Reserved(rpc ndr.Invoker) (err error) {
	req := &opnum222ReservedRequest{}
	var resp opnum222ReservedResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("Opnum222Reserved: %w", err)
		return
	}
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("Opnum222Reserved failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
