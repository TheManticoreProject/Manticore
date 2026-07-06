package functions

import (
	"fmt"

	dimsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/8f09f000-b7ed-11ce-bbd2-00001a181cad/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rRasAdminPortResetRequest carries the [in] parameters of RRasAdminPortReset.
type rRasAdminPortResetRequest struct {
	HPort ndr.DWORD
}

func (*rRasAdminPortResetRequest) Opnum() uint16 { return dimsvc.OpnumRRasAdminPortReset }

// rRasAdminPortResetResponse carries the [out] parameters and return value of RRasAdminPortReset.
type rRasAdminPortResetResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RRasAdminPortReset calls RRasAdminPortReset (opnum 7) ([MS-RRASM] — verify the parameter
// modeling and status handling).
func RRasAdminPortReset(rpc ndr.Invoker, hPort ndr.DWORD) (err error) {
	req := &rRasAdminPortResetRequest{
		HPort: hPort,
	}
	var resp rRasAdminPortResetResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRasAdminPortReset: %w", err)
		return
	}
	if uint32(resp.Status) != dimsvc.StatusSuccess {
		err = fmt.Errorf("RRasAdminPortReset failed: %s", dimsvc.StatusString(uint32(resp.Status)))
	}
	return
}
