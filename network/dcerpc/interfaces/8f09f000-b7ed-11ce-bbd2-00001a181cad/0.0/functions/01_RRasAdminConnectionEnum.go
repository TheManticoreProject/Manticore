package functions

// IDL source: [MS-RRASM] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rrasm/8e6e89fb-9c80-4a9a-a222-d7d8948244bb
// A fetched copy is kept at ms-rrasm.idl in the interface directory.

import (
	"fmt"

	dimsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/8f09f000-b7ed-11ce-bbd2-00001a181cad/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrrasm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rrasm"
)

// rRasAdminConnectionEnumRequest carries the [in] parameters of RRasAdminConnectionEnum.
type rRasAdminConnectionEnumRequest struct {
	DwLevel                 ndr.DWORD
	PInfoStruct             msrrasm.DIM_INFORMATION_CONTAINER
	DwPreferedMaximumLength ndr.DWORD
	LpdwResumeHandle        *ndr.DWORD `ndr:"unique"`
}

func (*rRasAdminConnectionEnumRequest) Opnum() uint16 { return dimsvc.OpnumRRasAdminConnectionEnum }

// rRasAdminConnectionEnumResponse carries the [out] parameters and return value of RRasAdminConnectionEnum.
type rRasAdminConnectionEnumResponse struct {
	PInfoStruct      msrrasm.DIM_INFORMATION_CONTAINER
	LpdwEntriesRead  ndr.DWORD
	LpdwTotalEntries ndr.DWORD
	LpdwResumeHandle *ndr.DWORD `ndr:"unique"`
	Status           ndr.DWORD  `ndr:"retval"`
}

// RRasAdminConnectionEnum calls RRasAdminConnectionEnum (opnum 1) ([MS-RRASM] — verify the parameter
// modeling and status handling).
func RRasAdminConnectionEnum(rpc ndr.Invoker, dwLevel ndr.DWORD, pInfoStruct msrrasm.DIM_INFORMATION_CONTAINER, dwPreferedMaximumLength ndr.DWORD, lpdwResumeHandle *ndr.DWORD) (PInfoStruct msrrasm.DIM_INFORMATION_CONTAINER, LpdwEntriesRead ndr.DWORD, LpdwTotalEntries ndr.DWORD, LpdwResumeHandle *ndr.DWORD, err error) {
	req := &rRasAdminConnectionEnumRequest{
		DwLevel:                 dwLevel,
		PInfoStruct:             pInfoStruct,
		DwPreferedMaximumLength: dwPreferedMaximumLength,
		LpdwResumeHandle:        lpdwResumeHandle,
	}
	var resp rRasAdminConnectionEnumResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRasAdminConnectionEnum: %w", err)
		return
	}
	PInfoStruct = resp.PInfoStruct
	LpdwEntriesRead = resp.LpdwEntriesRead
	LpdwTotalEntries = resp.LpdwTotalEntries
	LpdwResumeHandle = resp.LpdwResumeHandle
	if uint32(resp.Status) != dimsvc.StatusSuccess {
		err = fmt.Errorf("RRasAdminConnectionEnum failed: %s", dimsvc.StatusString(uint32(resp.Status)))
	}
	return
}
