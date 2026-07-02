package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// opnum191ReservedRequest carries the [in] parameters of Opnum191Reserved.
type opnum191ReservedRequest struct {
}

func (*opnum191ReservedRequest) Opnum() uint16 { return clusapi.OpnumOpnum191Reserved }

// opnum191ReservedResponse carries the [out] parameters and return value of Opnum191Reserved.
type opnum191ReservedResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// Opnum191Reserved calls Opnum191Reserved (opnum 191) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func Opnum191Reserved(rpc ndr.Invoker) (err error) {
	req := &opnum191ReservedRequest{}
	var resp opnum191ReservedResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("Opnum191Reserved: %w", err)
		return
	}
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("Opnum191Reserved failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
