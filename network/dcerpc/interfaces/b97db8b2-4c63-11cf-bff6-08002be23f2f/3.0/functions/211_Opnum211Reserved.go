package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// opnum211ReservedRequest carries the [in] parameters of Opnum211Reserved.
type opnum211ReservedRequest struct {
}

func (*opnum211ReservedRequest) Opnum() uint16 { return clusapi.OpnumOpnum211Reserved }

// opnum211ReservedResponse carries the [out] parameters and return value of Opnum211Reserved.
type opnum211ReservedResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// Opnum211Reserved calls Opnum211Reserved (opnum 211) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func Opnum211Reserved(rpc ndr.Invoker) (err error) {
	req := &opnum211ReservedRequest{}
	var resp opnum211ReservedResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("Opnum211Reserved: %w", err)
		return
	}
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("Opnum211Reserved failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
