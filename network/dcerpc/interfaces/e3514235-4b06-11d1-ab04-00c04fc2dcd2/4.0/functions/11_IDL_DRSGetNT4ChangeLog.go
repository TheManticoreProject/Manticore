package functions

import (
	"fmt"

	drsuapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// iDL_DRSGetNT4ChangeLogRequest carries the [in] parameters of IDL_DRSGetNT4ChangeLog.
type iDL_DRSGetNT4ChangeLogRequest struct {
	HDrs        structures.DRS_HANDLE
	DwInVersion ndr.DWORD
	PmsgIn      structures.DRS_MSG_NT4_CHGLOG_REQ
}

func (*iDL_DRSGetNT4ChangeLogRequest) Opnum() uint16 { return drsuapi.OpnumIDL_DRSGetNT4ChangeLog }

// iDL_DRSGetNT4ChangeLogResponse carries the [out] parameters and return value of IDL_DRSGetNT4ChangeLog.
type iDL_DRSGetNT4ChangeLogResponse struct {
	PdwOutVersion ndr.DWORD
	PmsgOut       structures.DRS_MSG_NT4_CHGLOG_REPLY
	Status        ndr.DWORD `ndr:"retval"`
}

// IDL_DRSGetNT4ChangeLog calls IDL_DRSGetNT4ChangeLog (opnum 11) ([MS-DRSR] — verify the parameter
// modeling and status handling).
func IDL_DRSGetNT4ChangeLog(rpc ndr.Invoker, hDrs structures.DRS_HANDLE, dwInVersion ndr.DWORD, pmsgIn structures.DRS_MSG_NT4_CHGLOG_REQ) (PdwOutVersion ndr.DWORD, PmsgOut structures.DRS_MSG_NT4_CHGLOG_REPLY, err error) {
	req := &iDL_DRSGetNT4ChangeLogRequest{
		HDrs:        hDrs,
		DwInVersion: dwInVersion,
		PmsgIn:      pmsgIn,
	}
	var resp iDL_DRSGetNT4ChangeLogResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("IDL_DRSGetNT4ChangeLog: %w", err)
		return
	}
	PdwOutVersion = resp.PdwOutVersion
	PmsgOut = resp.PmsgOut
	if uint32(resp.Status) != drsuapi.StatusSuccess {
		err = fmt.Errorf("IDL_DRSGetNT4ChangeLog failed: %s", drsuapi.StatusString(uint32(resp.Status)))
	}
	return
}
