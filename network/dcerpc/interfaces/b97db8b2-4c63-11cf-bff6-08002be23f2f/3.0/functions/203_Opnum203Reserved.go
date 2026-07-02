package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// opnum203ReservedRequest carries the [in] parameters of Opnum203Reserved.
type opnum203ReservedRequest struct {
}

func (*opnum203ReservedRequest) Opnum() uint16 { return clusapi.OpnumOpnum203Reserved }

// opnum203ReservedResponse carries the [out] parameters and return value of Opnum203Reserved.
type opnum203ReservedResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// Opnum203Reserved calls Opnum203Reserved (opnum 203) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func Opnum203Reserved(rpc ndr.Invoker) (err error) {
	req := &opnum203ReservedRequest{}
	var resp opnum203ReservedResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("Opnum203Reserved: %w", err)
		return
	}
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("Opnum203Reserved failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
