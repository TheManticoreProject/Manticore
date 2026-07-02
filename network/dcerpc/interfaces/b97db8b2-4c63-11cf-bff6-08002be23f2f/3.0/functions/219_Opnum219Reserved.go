package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// opnum219ReservedRequest carries the [in] parameters of Opnum219Reserved.
type opnum219ReservedRequest struct {
}

func (*opnum219ReservedRequest) Opnum() uint16 { return clusapi.OpnumOpnum219Reserved }

// opnum219ReservedResponse carries the [out] parameters and return value of Opnum219Reserved.
type opnum219ReservedResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// Opnum219Reserved calls Opnum219Reserved (opnum 219) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func Opnum219Reserved(rpc ndr.Invoker) (err error) {
	req := &opnum219ReservedRequest{}
	var resp opnum219ReservedResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("Opnum219Reserved: %w", err)
		return
	}
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("Opnum219Reserved failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
