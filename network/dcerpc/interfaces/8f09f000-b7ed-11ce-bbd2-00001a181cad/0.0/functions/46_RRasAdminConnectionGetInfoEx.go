package functions

import (
	"fmt"

	dimsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/8f09f000-b7ed-11ce-bbd2-00001a181cad/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrrasm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rrasm"
)

// rRasAdminConnectionGetInfoExRequest carries the [in] parameters of RRasAdminConnectionGetInfoEx.
type rRasAdminConnectionGetInfoExRequest struct {
	HDimConnection ndr.DWORD
	ObjectHeader   msrrasm.MPRAPI_OBJECT_HEADER_IDL
}

func (*rRasAdminConnectionGetInfoExRequest) Opnum() uint16 {
	return dimsvc.OpnumRRasAdminConnectionGetInfoEx
}

// rRasAdminConnectionGetInfoExResponse carries the [out] parameters and return value of RRasAdminConnectionGetInfoEx.
type rRasAdminConnectionGetInfoExResponse struct {
	PRasConnection msrrasm.PRAS_CONNECTION_EX_IDL
	Status         ndr.DWORD `ndr:"retval"`
}

// RRasAdminConnectionGetInfoEx calls RRasAdminConnectionGetInfoEx (opnum 46) ([MS-RRASM] — verify the parameter
// modeling and status handling).
func RRasAdminConnectionGetInfoEx(rpc ndr.Invoker, hDimConnection ndr.DWORD, objectHeader msrrasm.MPRAPI_OBJECT_HEADER_IDL) (PRasConnection msrrasm.PRAS_CONNECTION_EX_IDL, err error) {
	req := &rRasAdminConnectionGetInfoExRequest{
		HDimConnection: hDimConnection,
		ObjectHeader:   objectHeader,
	}
	var resp rRasAdminConnectionGetInfoExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRasAdminConnectionGetInfoEx: %w", err)
		return
	}
	PRasConnection = resp.PRasConnection
	if uint32(resp.Status) != dimsvc.StatusSuccess {
		err = fmt.Errorf("RRasAdminConnectionGetInfoEx failed: %s", dimsvc.StatusString(uint32(resp.Status)))
	}
	return
}
