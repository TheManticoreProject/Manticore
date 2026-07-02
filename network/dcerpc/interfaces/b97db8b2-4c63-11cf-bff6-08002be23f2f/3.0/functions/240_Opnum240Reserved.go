package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// opnum240ReservedRequest carries the [in] parameters of Opnum240Reserved.
type opnum240ReservedRequest struct {
}

func (*opnum240ReservedRequest) Opnum() uint16 { return clusapi.OpnumOpnum240Reserved }

// opnum240ReservedResponse carries the [out] parameters and return value of Opnum240Reserved.
type opnum240ReservedResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// Opnum240Reserved calls Opnum240Reserved (opnum 240) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func Opnum240Reserved(rpc ndr.Invoker) (err error) {
	req := &opnum240ReservedRequest{}
	var resp opnum240ReservedResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("Opnum240Reserved: %w", err)
		return
	}
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("Opnum240Reserved failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
