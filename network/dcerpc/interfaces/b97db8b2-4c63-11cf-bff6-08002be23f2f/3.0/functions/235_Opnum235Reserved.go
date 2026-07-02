package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// opnum235ReservedRequest carries the [in] parameters of Opnum235Reserved.
type opnum235ReservedRequest struct {
}

func (*opnum235ReservedRequest) Opnum() uint16 { return clusapi.OpnumOpnum235Reserved }

// opnum235ReservedResponse carries the [out] parameters and return value of Opnum235Reserved.
type opnum235ReservedResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// Opnum235Reserved calls Opnum235Reserved (opnum 235) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func Opnum235Reserved(rpc ndr.Invoker) (err error) {
	req := &opnum235ReservedRequest{}
	var resp opnum235ReservedResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("Opnum235Reserved: %w", err)
		return
	}
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("Opnum235Reserved failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
