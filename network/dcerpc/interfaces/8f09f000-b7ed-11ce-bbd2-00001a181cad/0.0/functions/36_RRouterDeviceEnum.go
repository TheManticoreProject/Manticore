package functions

import (
	"fmt"

	dimsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/8f09f000-b7ed-11ce-bbd2-00001a181cad/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrrasm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rrasm"
)

// rRouterDeviceEnumRequest carries the [in] parameters of RRouterDeviceEnum.
type rRouterDeviceEnumRequest struct {
	DwLevel          ndr.DWORD
	PInfoStruct      msrrasm.DIM_INFORMATION_CONTAINER
	LpdwTotalEntries ndr.DWORD
}

func (*rRouterDeviceEnumRequest) Opnum() uint16 { return dimsvc.OpnumRRouterDeviceEnum }

// rRouterDeviceEnumResponse carries the [out] parameters and return value of RRouterDeviceEnum.
type rRouterDeviceEnumResponse struct {
	PInfoStruct      msrrasm.DIM_INFORMATION_CONTAINER
	LpdwTotalEntries ndr.DWORD
	Status           ndr.DWORD `ndr:"retval"`
}

// RRouterDeviceEnum calls RRouterDeviceEnum (opnum 36) ([MS-RRASM] — verify the parameter
// modeling and status handling).
func RRouterDeviceEnum(rpc ndr.Invoker, dwLevel ndr.DWORD, pInfoStruct msrrasm.DIM_INFORMATION_CONTAINER, lpdwTotalEntries ndr.DWORD) (PInfoStruct msrrasm.DIM_INFORMATION_CONTAINER, LpdwTotalEntries ndr.DWORD, err error) {
	req := &rRouterDeviceEnumRequest{
		DwLevel:          dwLevel,
		PInfoStruct:      pInfoStruct,
		LpdwTotalEntries: lpdwTotalEntries,
	}
	var resp rRouterDeviceEnumResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRouterDeviceEnum: %w", err)
		return
	}
	PInfoStruct = resp.PInfoStruct
	LpdwTotalEntries = resp.LpdwTotalEntries
	if uint32(resp.Status) != dimsvc.StatusSuccess {
		err = fmt.Errorf("RRouterDeviceEnum failed: %s", dimsvc.StatusString(uint32(resp.Status)))
	}
	return
}
