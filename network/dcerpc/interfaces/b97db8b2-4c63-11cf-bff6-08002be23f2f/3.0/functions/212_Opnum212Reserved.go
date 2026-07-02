package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// opnum212ReservedRequest carries the [in] parameters of Opnum212Reserved.
type opnum212ReservedRequest struct {
}

func (*opnum212ReservedRequest) Opnum() uint16 { return clusapi.OpnumOpnum212Reserved }

// opnum212ReservedResponse carries the [out] parameters and return value of Opnum212Reserved.
type opnum212ReservedResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// Opnum212Reserved calls Opnum212Reserved (opnum 212) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func Opnum212Reserved(rpc ndr.Invoker) (err error) {
	req := &opnum212ReservedRequest{}
	var resp opnum212ReservedResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("Opnum212Reserved: %w", err)
		return
	}
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("Opnum212Reserved failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
