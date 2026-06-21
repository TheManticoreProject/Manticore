package functions

import (
	"fmt"

	drsuapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// iDL_DRSVerifyNamesRequest carries the [in] parameters of IDL_DRSVerifyNames.
type iDL_DRSVerifyNamesRequest struct {
	HDrs        structures.DRS_HANDLE
	DwInVersion ndr.DWORD
	PmsgIn      structures.DRS_MSG_VERIFYREQ
}

func (*iDL_DRSVerifyNamesRequest) Opnum() uint16 { return drsuapi.OpnumIDL_DRSVerifyNames }

// iDL_DRSVerifyNamesResponse carries the [out] parameters and return value of IDL_DRSVerifyNames.
type iDL_DRSVerifyNamesResponse struct {
	PdwOutVersion ndr.DWORD
	PmsgOut       structures.DRS_MSG_VERIFYREPLY
	Status        ndr.DWORD `ndr:"retval"`
}

// IDL_DRSVerifyNames calls IDL_DRSVerifyNames (opnum 8) ([MS-DRSR] — verify the parameter
// modeling and status handling).
func IDL_DRSVerifyNames(rpc ndr.Invoker, hDrs structures.DRS_HANDLE, dwInVersion ndr.DWORD, pmsgIn structures.DRS_MSG_VERIFYREQ) (PdwOutVersion ndr.DWORD, PmsgOut structures.DRS_MSG_VERIFYREPLY, err error) {
	req := &iDL_DRSVerifyNamesRequest{
		HDrs:        hDrs,
		DwInVersion: dwInVersion,
		PmsgIn:      pmsgIn,
	}
	var resp iDL_DRSVerifyNamesResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("IDL_DRSVerifyNames: %w", err)
		return
	}
	PdwOutVersion = resp.PdwOutVersion
	PmsgOut = resp.PmsgOut
	if uint32(resp.Status) != drsuapi.StatusSuccess {
		err = fmt.Errorf("IDL_DRSVerifyNames failed: %s", drsuapi.StatusString(uint32(resp.Status)))
	}
	return
}
