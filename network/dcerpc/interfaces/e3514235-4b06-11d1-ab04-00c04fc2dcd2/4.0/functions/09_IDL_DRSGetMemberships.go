package functions

import (
	"fmt"

	drsuapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// iDL_DRSGetMembershipsRequest carries the [in] parameters of IDL_DRSGetMemberships.
type iDL_DRSGetMembershipsRequest struct {
	HDrs        structures.DRS_HANDLE
	DwInVersion ndr.DWORD
	PmsgIn      structures.DRS_MSG_REVMEMB_REQ
}

func (*iDL_DRSGetMembershipsRequest) Opnum() uint16 { return drsuapi.OpnumIDL_DRSGetMemberships }

// iDL_DRSGetMembershipsResponse carries the [out] parameters and return value of IDL_DRSGetMemberships.
type iDL_DRSGetMembershipsResponse struct {
	PdwOutVersion ndr.DWORD
	PmsgOut       structures.DRS_MSG_REVMEMB_REPLY
	Status        ndr.DWORD `ndr:"retval"`
}

// IDL_DRSGetMemberships calls IDL_DRSGetMemberships (opnum 9) ([MS-DRSR] — verify the parameter
// modeling and status handling).
func IDL_DRSGetMemberships(rpc ndr.Invoker, hDrs structures.DRS_HANDLE, dwInVersion ndr.DWORD, pmsgIn structures.DRS_MSG_REVMEMB_REQ) (PdwOutVersion ndr.DWORD, PmsgOut structures.DRS_MSG_REVMEMB_REPLY, err error) {
	req := &iDL_DRSGetMembershipsRequest{
		HDrs:        hDrs,
		DwInVersion: dwInVersion,
		PmsgIn:      pmsgIn,
	}
	var resp iDL_DRSGetMembershipsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("IDL_DRSGetMemberships: %w", err)
		return
	}
	PdwOutVersion = resp.PdwOutVersion
	PmsgOut = resp.PmsgOut
	if uint32(resp.Status) != drsuapi.StatusSuccess {
		err = fmt.Errorf("IDL_DRSGetMemberships failed: %s", drsuapi.StatusString(uint32(resp.Status)))
	}
	return
}
