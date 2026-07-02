package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// opnum216ReservedRequest carries the [in] parameters of Opnum216Reserved.
type opnum216ReservedRequest struct {
}

func (*opnum216ReservedRequest) Opnum() uint16 { return clusapi.OpnumOpnum216Reserved }

// opnum216ReservedResponse carries the [out] parameters and return value of Opnum216Reserved.
type opnum216ReservedResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// Opnum216Reserved calls Opnum216Reserved (opnum 216) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func Opnum216Reserved(rpc ndr.Invoker) (err error) {
	req := &opnum216ReservedRequest{}
	var resp opnum216ReservedResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("Opnum216Reserved: %w", err)
		return
	}
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("Opnum216Reserved failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
