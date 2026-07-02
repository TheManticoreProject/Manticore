package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// opnum209ReservedRequest carries the [in] parameters of Opnum209Reserved.
type opnum209ReservedRequest struct {
}

func (*opnum209ReservedRequest) Opnum() uint16 { return clusapi.OpnumOpnum209Reserved }

// opnum209ReservedResponse carries the [out] parameters and return value of Opnum209Reserved.
type opnum209ReservedResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// Opnum209Reserved calls Opnum209Reserved (opnum 209) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func Opnum209Reserved(rpc ndr.Invoker) (err error) {
	req := &opnum209ReservedRequest{}
	var resp opnum209ReservedResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("Opnum209Reserved: %w", err)
		return
	}
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("Opnum209Reserved failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
