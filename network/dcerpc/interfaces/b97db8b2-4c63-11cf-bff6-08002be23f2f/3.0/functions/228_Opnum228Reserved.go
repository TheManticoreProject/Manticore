package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// opnum228ReservedRequest carries the [in] parameters of Opnum228Reserved.
type opnum228ReservedRequest struct {
}

func (*opnum228ReservedRequest) Opnum() uint16 { return clusapi.OpnumOpnum228Reserved }

// opnum228ReservedResponse carries the [out] parameters and return value of Opnum228Reserved.
type opnum228ReservedResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// Opnum228Reserved calls Opnum228Reserved (opnum 228) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func Opnum228Reserved(rpc ndr.Invoker) (err error) {
	req := &opnum228ReservedRequest{}
	var resp opnum228ReservedResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("Opnum228Reserved: %w", err)
		return
	}
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("Opnum228Reserved failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
