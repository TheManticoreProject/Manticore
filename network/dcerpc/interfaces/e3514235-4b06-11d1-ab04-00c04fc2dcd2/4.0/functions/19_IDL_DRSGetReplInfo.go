package functions

import (
	"fmt"

	drsuapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// iDL_DRSGetReplInfoRequest carries the [in] parameters of IDL_DRSGetReplInfo.
type iDL_DRSGetReplInfoRequest struct {
	HDrs        structures.DRS_HANDLE
	DwInVersion ndr.DWORD
	PmsgIn      structures.DRS_MSG_GETREPLINFO_REQ
}

func (*iDL_DRSGetReplInfoRequest) Opnum() uint16 { return drsuapi.OpnumIDL_DRSGetReplInfo }

// iDL_DRSGetReplInfoResponse carries the [out] parameters and return value of IDL_DRSGetReplInfo.
type iDL_DRSGetReplInfoResponse struct {
	PdwOutVersion ndr.DWORD
	PmsgOut       structures.DRS_MSG_GETREPLINFO_REPLY
	Status        ndr.DWORD `ndr:"retval"`
}

// IDL_DRSGetReplInfo calls IDL_DRSGetReplInfo (opnum 19) ([MS-DRSR] — verify the parameter
// modeling and status handling).
func IDL_DRSGetReplInfo(rpc ndr.Invoker, hDrs structures.DRS_HANDLE, dwInVersion ndr.DWORD, pmsgIn structures.DRS_MSG_GETREPLINFO_REQ) (PdwOutVersion ndr.DWORD, PmsgOut structures.DRS_MSG_GETREPLINFO_REPLY, err error) {
	req := &iDL_DRSGetReplInfoRequest{
		HDrs:        hDrs,
		DwInVersion: dwInVersion,
		PmsgIn:      pmsgIn,
	}
	var resp iDL_DRSGetReplInfoResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("IDL_DRSGetReplInfo: %w", err)
		return
	}
	PdwOutVersion = resp.PdwOutVersion
	PmsgOut = resp.PmsgOut
	if uint32(resp.Status) != drsuapi.StatusSuccess {
		err = fmt.Errorf("IDL_DRSGetReplInfo failed: %s", drsuapi.StatusString(uint32(resp.Status)))
	}
	return
}
