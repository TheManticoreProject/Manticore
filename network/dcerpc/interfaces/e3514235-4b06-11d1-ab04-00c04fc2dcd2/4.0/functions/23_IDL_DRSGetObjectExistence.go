package functions

import (
	"fmt"

	drsuapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// iDL_DRSGetObjectExistenceRequest carries the [in] parameters of IDL_DRSGetObjectExistence.
type iDL_DRSGetObjectExistenceRequest struct {
	HDrs        structures.DRS_HANDLE
	DwInVersion ndr.DWORD
	PmsgIn      structures.DRS_MSG_EXISTREQ
}

func (*iDL_DRSGetObjectExistenceRequest) Opnum() uint16 {
	return drsuapi.OpnumIDL_DRSGetObjectExistence
}

// iDL_DRSGetObjectExistenceResponse carries the [out] parameters and return value of IDL_DRSGetObjectExistence.
type iDL_DRSGetObjectExistenceResponse struct {
	PdwOutVersion ndr.DWORD
	PmsgOut       structures.DRS_MSG_EXISTREPLY
	Status        ndr.DWORD `ndr:"retval"`
}

// IDL_DRSGetObjectExistence calls IDL_DRSGetObjectExistence (opnum 23) ([MS-DRSR] — verify the parameter
// modeling and status handling).
func IDL_DRSGetObjectExistence(rpc ndr.Invoker, hDrs structures.DRS_HANDLE, dwInVersion ndr.DWORD, pmsgIn structures.DRS_MSG_EXISTREQ) (PdwOutVersion ndr.DWORD, PmsgOut structures.DRS_MSG_EXISTREPLY, err error) {
	req := &iDL_DRSGetObjectExistenceRequest{
		HDrs:        hDrs,
		DwInVersion: dwInVersion,
		PmsgIn:      pmsgIn,
	}
	var resp iDL_DRSGetObjectExistenceResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("IDL_DRSGetObjectExistence: %w", err)
		return
	}
	PdwOutVersion = resp.PdwOutVersion
	PmsgOut = resp.PmsgOut
	if uint32(resp.Status) != drsuapi.StatusSuccess {
		err = fmt.Errorf("IDL_DRSGetObjectExistence failed: %s", drsuapi.StatusString(uint32(resp.Status)))
	}
	return
}
