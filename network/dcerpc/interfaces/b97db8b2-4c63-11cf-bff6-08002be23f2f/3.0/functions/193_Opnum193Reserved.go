package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// opnum193ReservedRequest carries the [in] parameters of Opnum193Reserved.
type opnum193ReservedRequest struct {
}

func (*opnum193ReservedRequest) Opnum() uint16 { return clusapi.OpnumOpnum193Reserved }

// opnum193ReservedResponse carries the [out] parameters and return value of Opnum193Reserved.
type opnum193ReservedResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// Opnum193Reserved calls Opnum193Reserved (opnum 193) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func Opnum193Reserved(rpc ndr.Invoker) (err error) {
	req := &opnum193ReservedRequest{}
	var resp opnum193ReservedResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("Opnum193Reserved: %w", err)
		return
	}
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("Opnum193Reserved failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
