package functions

// IDL source: [MS-DRSR] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-drsr/3f5d9495-9563-44de-876a-ce6f880e3fb2
// A fetched copy is kept at ms-drsr.idl in the interface directory.

import (
	"fmt"

	drsuapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdrsr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-drsr"
)

// iDL_DRSGetNCChangesRequest carries the [in] parameters of IDL_DRSGetNCChanges.
type iDL_DRSGetNCChangesRequest struct {
	HDrs        msdrsr.DRS_HANDLE
	DwInVersion ndr.DWORD
	PmsgIn      msdrsr.DRS_MSG_GETCHGREQ
}

func (*iDL_DRSGetNCChangesRequest) Opnum() uint16 { return drsuapi.OpnumIDL_DRSGetNCChanges }

// iDL_DRSGetNCChangesResponse carries the [out] parameters and return value of IDL_DRSGetNCChanges.
type iDL_DRSGetNCChangesResponse struct {
	PdwOutVersion ndr.DWORD
	PmsgOut       msdrsr.DRS_MSG_GETCHGREPLY
	Status        ndr.DWORD `ndr:"retval"`
}

// IDL_DRSGetNCChanges calls IDL_DRSGetNCChanges (opnum 3) ([MS-DRSR] — verify the parameter
// modeling and status handling).
func IDL_DRSGetNCChanges(rpc ndr.Invoker, hDrs msdrsr.DRS_HANDLE, dwInVersion ndr.DWORD, pmsgIn msdrsr.DRS_MSG_GETCHGREQ) (PdwOutVersion ndr.DWORD, PmsgOut msdrsr.DRS_MSG_GETCHGREPLY, err error) {
	req := &iDL_DRSGetNCChangesRequest{
		HDrs:        hDrs,
		DwInVersion: dwInVersion,
		PmsgIn:      pmsgIn,
	}
	var resp iDL_DRSGetNCChangesResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("IDL_DRSGetNCChanges: %w", err)
		return
	}
	PdwOutVersion = resp.PdwOutVersion
	PmsgOut = resp.PmsgOut
	if uint32(resp.Status) != drsuapi.StatusSuccess {
		err = fmt.Errorf("IDL_DRSGetNCChanges failed: %s", drsuapi.StatusString(uint32(resp.Status)))
	}
	return
}
