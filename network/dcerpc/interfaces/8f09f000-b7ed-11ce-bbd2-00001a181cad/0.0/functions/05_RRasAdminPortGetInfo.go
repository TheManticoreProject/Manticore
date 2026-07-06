package functions

import (
	"fmt"

	dimsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/8f09f000-b7ed-11ce-bbd2-00001a181cad/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrrasm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rrasm"
)

// rRasAdminPortGetInfoRequest carries the [in] parameters of RRasAdminPortGetInfo.
type rRasAdminPortGetInfoRequest struct {
	DwLevel ndr.DWORD
	HPort   ndr.DWORD
}

func (*rRasAdminPortGetInfoRequest) Opnum() uint16 { return dimsvc.OpnumRRasAdminPortGetInfo }

// rRasAdminPortGetInfoResponse carries the [out] parameters and return value of RRasAdminPortGetInfo.
type rRasAdminPortGetInfoResponse struct {
	PInfoStruct msrrasm.DIM_INFORMATION_CONTAINER
	Status      ndr.DWORD `ndr:"retval"`
}

// RRasAdminPortGetInfo calls RRasAdminPortGetInfo (opnum 5) ([MS-RRASM] — verify the parameter
// modeling and status handling).
func RRasAdminPortGetInfo(rpc ndr.Invoker, dwLevel ndr.DWORD, hPort ndr.DWORD) (PInfoStruct msrrasm.DIM_INFORMATION_CONTAINER, err error) {
	req := &rRasAdminPortGetInfoRequest{
		DwLevel: dwLevel,
		HPort:   hPort,
	}
	var resp rRasAdminPortGetInfoResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRasAdminPortGetInfo: %w", err)
		return
	}
	PInfoStruct = resp.PInfoStruct
	if uint32(resp.Status) != dimsvc.StatusSuccess {
		err = fmt.Errorf("RRasAdminPortGetInfo failed: %s", dimsvc.StatusString(uint32(resp.Status)))
	}
	return
}
