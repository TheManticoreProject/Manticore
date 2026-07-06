package functions

import (
	"fmt"

	dimsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/8f09f000-b7ed-11ce-bbd2-00001a181cad/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrrasm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rrasm"
)

// rMprAdminServerGetInfoRequest carries the [in] parameters of RMprAdminServerGetInfo.
type rMprAdminServerGetInfoRequest struct {
	DwLevel ndr.DWORD
}

func (*rMprAdminServerGetInfoRequest) Opnum() uint16 { return dimsvc.OpnumRMprAdminServerGetInfo }

// rMprAdminServerGetInfoResponse carries the [out] parameters and return value of RMprAdminServerGetInfo.
type rMprAdminServerGetInfoResponse struct {
	PInfoStruct msrrasm.DIM_INFORMATION_CONTAINER
	Status      ndr.DWORD `ndr:"retval"`
}

// RMprAdminServerGetInfo calls RMprAdminServerGetInfo (opnum 0) ([MS-RRASM] — verify the parameter
// modeling and status handling).
func RMprAdminServerGetInfo(rpc ndr.Invoker, dwLevel ndr.DWORD) (PInfoStruct msrrasm.DIM_INFORMATION_CONTAINER, err error) {
	req := &rMprAdminServerGetInfoRequest{
		DwLevel: dwLevel,
	}
	var resp rMprAdminServerGetInfoResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RMprAdminServerGetInfo: %w", err)
		return
	}
	PInfoStruct = resp.PInfoStruct
	if uint32(resp.Status) != dimsvc.StatusSuccess {
		err = fmt.Errorf("RMprAdminServerGetInfo failed: %s", dimsvc.StatusString(uint32(resp.Status)))
	}
	return
}
