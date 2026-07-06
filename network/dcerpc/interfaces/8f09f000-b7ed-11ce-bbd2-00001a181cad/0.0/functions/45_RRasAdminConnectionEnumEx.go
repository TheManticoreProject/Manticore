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

// rRasAdminConnectionEnumExRequest carries the [in] parameters of RRasAdminConnectionEnumEx.
type rRasAdminConnectionEnumExRequest struct {
	ObjectHeader     msrrasm.MPRAPI_OBJECT_HEADER_IDL
	DwPreferedMaxLen ndr.DWORD
	LpdwResumeHandle *ndr.DWORD `ndr:"unique"`
}

func (*rRasAdminConnectionEnumExRequest) Opnum() uint16 { return dimsvc.OpnumRRasAdminConnectionEnumEx }

// rRasAdminConnectionEnumExResponse carries the [out] parameters and return value of RRasAdminConnectionEnumEx.
type rRasAdminConnectionEnumExResponse struct {
	LpdwEntriesRead     ndr.DWORD
	LpdNumTotalElements ndr.DWORD
	PRasConections      []msrrasm.PRAS_CONNECTION_EX_IDL `ndr:"ref,conformant"`
	LpdwResumeHandle    *ndr.DWORD                       `ndr:"unique"`
	Status              ndr.DWORD                        `ndr:"retval"`
}

// RRasAdminConnectionEnumEx calls RRasAdminConnectionEnumEx (opnum 45) ([MS-RRASM] — verify the parameter
// modeling and status handling).
func RRasAdminConnectionEnumEx(rpc ndr.Invoker, objectHeader msrrasm.MPRAPI_OBJECT_HEADER_IDL, dwPreferedMaxLen ndr.DWORD, lpdwResumeHandle *ndr.DWORD) (LpdwEntriesRead ndr.DWORD, LpdNumTotalElements ndr.DWORD, PRasConections []msrrasm.PRAS_CONNECTION_EX_IDL, LpdwResumeHandle *ndr.DWORD, err error) {
	req := &rRasAdminConnectionEnumExRequest{
		ObjectHeader:     objectHeader,
		DwPreferedMaxLen: dwPreferedMaxLen,
		LpdwResumeHandle: lpdwResumeHandle,
	}
	var resp rRasAdminConnectionEnumExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRasAdminConnectionEnumEx: %w", err)
		return
	}
	LpdwEntriesRead = resp.LpdwEntriesRead
	LpdNumTotalElements = resp.LpdNumTotalElements
	PRasConections = resp.PRasConections
	LpdwResumeHandle = resp.LpdwResumeHandle
	if uint32(resp.Status) != dimsvc.StatusSuccess {
		err = fmt.Errorf("RRasAdminConnectionEnumEx failed: %s", dimsvc.StatusString(uint32(resp.Status)))
	}
	return
}
