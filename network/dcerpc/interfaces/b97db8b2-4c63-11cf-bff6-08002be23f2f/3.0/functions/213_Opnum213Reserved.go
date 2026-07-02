package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// opnum213ReservedRequest carries the [in] parameters of Opnum213Reserved.
type opnum213ReservedRequest struct {
}

func (*opnum213ReservedRequest) Opnum() uint16 { return clusapi.OpnumOpnum213Reserved }

// opnum213ReservedResponse carries the [out] parameters and return value of Opnum213Reserved.
type opnum213ReservedResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// Opnum213Reserved calls Opnum213Reserved (opnum 213) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func Opnum213Reserved(rpc ndr.Invoker) (err error) {
	req := &opnum213ReservedRequest{}
	var resp opnum213ReservedResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("Opnum213Reserved: %w", err)
		return
	}
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("Opnum213Reserved failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
