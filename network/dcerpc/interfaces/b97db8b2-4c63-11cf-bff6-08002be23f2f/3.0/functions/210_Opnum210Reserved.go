package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// opnum210ReservedRequest carries the [in] parameters of Opnum210Reserved.
type opnum210ReservedRequest struct {
}

func (*opnum210ReservedRequest) Opnum() uint16 { return clusapi.OpnumOpnum210Reserved }

// opnum210ReservedResponse carries the [out] parameters and return value of Opnum210Reserved.
type opnum210ReservedResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// Opnum210Reserved calls Opnum210Reserved (opnum 210) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func Opnum210Reserved(rpc ndr.Invoker) (err error) {
	req := &opnum210ReservedRequest{}
	var resp opnum210ReservedResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("Opnum210Reserved: %w", err)
		return
	}
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("Opnum210Reserved failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
