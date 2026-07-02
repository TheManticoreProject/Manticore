package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// opnum245ReservedRequest carries the [in] parameters of Opnum245Reserved.
type opnum245ReservedRequest struct {
}

func (*opnum245ReservedRequest) Opnum() uint16 { return clusapi.OpnumOpnum245Reserved }

// opnum245ReservedResponse carries the [out] parameters and return value of Opnum245Reserved.
type opnum245ReservedResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// Opnum245Reserved calls Opnum245Reserved (opnum 245) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func Opnum245Reserved(rpc ndr.Invoker) (err error) {
	req := &opnum245ReservedRequest{}
	var resp opnum245ReservedResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("Opnum245Reserved: %w", err)
		return
	}
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("Opnum245Reserved failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
