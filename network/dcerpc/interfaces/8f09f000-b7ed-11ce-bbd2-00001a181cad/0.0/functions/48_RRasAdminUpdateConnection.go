package functions

import (
	"fmt"

	dimsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/8f09f000-b7ed-11ce-bbd2-00001a181cad/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrrasm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rrasm"
)

// rRasAdminUpdateConnectionRequest carries the [in] parameters of RRasAdminUpdateConnection.
type rRasAdminUpdateConnectionRequest struct {
	HDimConnection ndr.DWORD
	PServerConfig  msrrasm.PRAS_UPDATE_CONNECTION_IDL
}

func (*rRasAdminUpdateConnectionRequest) Opnum() uint16 { return dimsvc.OpnumRRasAdminUpdateConnection }

// rRasAdminUpdateConnectionResponse carries the [out] parameters and return value of RRasAdminUpdateConnection.
type rRasAdminUpdateConnectionResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RRasAdminUpdateConnection calls RRasAdminUpdateConnection (opnum 48) ([MS-RRASM] — verify the parameter
// modeling and status handling).
func RRasAdminUpdateConnection(rpc ndr.Invoker, hDimConnection ndr.DWORD, pServerConfig msrrasm.PRAS_UPDATE_CONNECTION_IDL) (err error) {
	req := &rRasAdminUpdateConnectionRequest{
		HDimConnection: hDimConnection,
		PServerConfig:  pServerConfig,
	}
	var resp rRasAdminUpdateConnectionResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRasAdminUpdateConnection: %w", err)
		return
	}
	if uint32(resp.Status) != dimsvc.StatusSuccess {
		err = fmt.Errorf("RRasAdminUpdateConnection failed: %s", dimsvc.StatusString(uint32(resp.Status)))
	}
	return
}
