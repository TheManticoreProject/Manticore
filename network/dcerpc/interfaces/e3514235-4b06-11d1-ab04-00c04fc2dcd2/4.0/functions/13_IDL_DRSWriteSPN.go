package functions

import (
	"fmt"

	drsuapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// iDL_DRSWriteSPNRequest carries the [in] parameters of IDL_DRSWriteSPN.
type iDL_DRSWriteSPNRequest struct {
	HDrs        structures.DRS_HANDLE
	DwInVersion ndr.DWORD
	PmsgIn      structures.DRS_MSG_SPNREQ
}

func (*iDL_DRSWriteSPNRequest) Opnum() uint16 { return drsuapi.OpnumIDL_DRSWriteSPN }

// iDL_DRSWriteSPNResponse carries the [out] parameters and return value of IDL_DRSWriteSPN.
type iDL_DRSWriteSPNResponse struct {
	PdwOutVersion ndr.DWORD
	PmsgOut       structures.DRS_MSG_SPNREPLY
	Status        ndr.DWORD `ndr:"retval"`
}

// IDL_DRSWriteSPN calls IDL_DRSWriteSPN (opnum 13) ([MS-DRSR] — verify the parameter
// modeling and status handling).
func IDL_DRSWriteSPN(rpc ndr.Invoker, hDrs structures.DRS_HANDLE, dwInVersion ndr.DWORD, pmsgIn structures.DRS_MSG_SPNREQ) (PdwOutVersion ndr.DWORD, PmsgOut structures.DRS_MSG_SPNREPLY, err error) {
	req := &iDL_DRSWriteSPNRequest{
		HDrs:        hDrs,
		DwInVersion: dwInVersion,
		PmsgIn:      pmsgIn,
	}
	var resp iDL_DRSWriteSPNResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("IDL_DRSWriteSPN: %w", err)
		return
	}
	PdwOutVersion = resp.PdwOutVersion
	PmsgOut = resp.PmsgOut
	if uint32(resp.Status) != drsuapi.StatusSuccess {
		err = fmt.Errorf("IDL_DRSWriteSPN failed: %s", drsuapi.StatusString(uint32(resp.Status)))
	}
	return
}
