package functions

import (
	"fmt"

	dimsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/8f09f000-b7ed-11ce-bbd2-00001a181cad/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrrasm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rrasm"
)

// rRasAdminConnectionGetInfoRequest carries the [in] parameters of RRasAdminConnectionGetInfo.
type rRasAdminConnectionGetInfoRequest struct {
	DwLevel        ndr.DWORD
	HDimConnection ndr.DWORD
}

func (*rRasAdminConnectionGetInfoRequest) Opnum() uint16 {
	return dimsvc.OpnumRRasAdminConnectionGetInfo
}

// rRasAdminConnectionGetInfoResponse carries the [out] parameters and return value of RRasAdminConnectionGetInfo.
type rRasAdminConnectionGetInfoResponse struct {
	PInfoStruct msrrasm.DIM_INFORMATION_CONTAINER
	Status      ndr.DWORD `ndr:"retval"`
}

// RRasAdminConnectionGetInfo calls RRasAdminConnectionGetInfo (opnum 2) ([MS-RRASM] — verify the parameter
// modeling and status handling).
func RRasAdminConnectionGetInfo(rpc ndr.Invoker, dwLevel ndr.DWORD, hDimConnection ndr.DWORD) (PInfoStruct msrrasm.DIM_INFORMATION_CONTAINER, err error) {
	req := &rRasAdminConnectionGetInfoRequest{
		DwLevel:        dwLevel,
		HDimConnection: hDimConnection,
	}
	var resp rRasAdminConnectionGetInfoResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRasAdminConnectionGetInfo: %w", err)
		return
	}
	PInfoStruct = resp.PInfoStruct
	if uint32(resp.Status) != dimsvc.StatusSuccess {
		err = fmt.Errorf("RRasAdminConnectionGetInfo failed: %s", dimsvc.StatusString(uint32(resp.Status)))
	}
	return
}
