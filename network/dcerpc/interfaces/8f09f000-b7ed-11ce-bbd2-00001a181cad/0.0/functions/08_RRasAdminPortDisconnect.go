package functions

import (
	"fmt"

	dimsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/8f09f000-b7ed-11ce-bbd2-00001a181cad/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rRasAdminPortDisconnectRequest carries the [in] parameters of RRasAdminPortDisconnect.
type rRasAdminPortDisconnectRequest struct {
	HPort ndr.DWORD
}

func (*rRasAdminPortDisconnectRequest) Opnum() uint16 { return dimsvc.OpnumRRasAdminPortDisconnect }

// rRasAdminPortDisconnectResponse carries the [out] parameters and return value of RRasAdminPortDisconnect.
type rRasAdminPortDisconnectResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RRasAdminPortDisconnect calls RRasAdminPortDisconnect (opnum 8) ([MS-RRASM] — verify the parameter
// modeling and status handling).
func RRasAdminPortDisconnect(rpc ndr.Invoker, hPort ndr.DWORD) (err error) {
	req := &rRasAdminPortDisconnectRequest{
		HPort: hPort,
	}
	var resp rRasAdminPortDisconnectResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRasAdminPortDisconnect: %w", err)
		return
	}
	if uint32(resp.Status) != dimsvc.StatusSuccess {
		err = fmt.Errorf("RRasAdminPortDisconnect failed: %s", dimsvc.StatusString(uint32(resp.Status)))
	}
	return
}
