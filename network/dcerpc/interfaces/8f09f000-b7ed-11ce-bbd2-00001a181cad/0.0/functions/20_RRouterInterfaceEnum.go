package functions

import (
	"fmt"

	dimsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/8f09f000-b7ed-11ce-bbd2-00001a181cad/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrrasm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rrasm"
)

// rRouterInterfaceEnumRequest carries the [in] parameters of RRouterInterfaceEnum.
type rRouterInterfaceEnumRequest struct {
	DwLevel                 ndr.DWORD
	PInfoStruct             msrrasm.DIM_INFORMATION_CONTAINER
	DwPreferedMaximumLength ndr.DWORD
	LpdwResumeHandle        *ndr.DWORD `ndr:"unique"`
}

func (*rRouterInterfaceEnumRequest) Opnum() uint16 { return dimsvc.OpnumRRouterInterfaceEnum }

// rRouterInterfaceEnumResponse carries the [out] parameters and return value of RRouterInterfaceEnum.
type rRouterInterfaceEnumResponse struct {
	PInfoStruct      msrrasm.DIM_INFORMATION_CONTAINER
	LpdwEntriesRead  ndr.DWORD
	LpdwTotalEntries ndr.DWORD
	LpdwResumeHandle *ndr.DWORD `ndr:"unique"`
	Status           ndr.DWORD  `ndr:"retval"`
}

// RRouterInterfaceEnum calls RRouterInterfaceEnum (opnum 20) ([MS-RRASM] — verify the parameter
// modeling and status handling).
func RRouterInterfaceEnum(rpc ndr.Invoker, dwLevel ndr.DWORD, pInfoStruct msrrasm.DIM_INFORMATION_CONTAINER, dwPreferedMaximumLength ndr.DWORD, lpdwResumeHandle *ndr.DWORD) (PInfoStruct msrrasm.DIM_INFORMATION_CONTAINER, LpdwEntriesRead ndr.DWORD, LpdwTotalEntries ndr.DWORD, LpdwResumeHandle *ndr.DWORD, err error) {
	req := &rRouterInterfaceEnumRequest{
		DwLevel:                 dwLevel,
		PInfoStruct:             pInfoStruct,
		DwPreferedMaximumLength: dwPreferedMaximumLength,
		LpdwResumeHandle:        lpdwResumeHandle,
	}
	var resp rRouterInterfaceEnumResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRouterInterfaceEnum: %w", err)
		return
	}
	PInfoStruct = resp.PInfoStruct
	LpdwEntriesRead = resp.LpdwEntriesRead
	LpdwTotalEntries = resp.LpdwTotalEntries
	LpdwResumeHandle = resp.LpdwResumeHandle
	if uint32(resp.Status) != dimsvc.StatusSuccess {
		err = fmt.Errorf("RRouterInterfaceEnum failed: %s", dimsvc.StatusString(uint32(resp.Status)))
	}
	return
}
