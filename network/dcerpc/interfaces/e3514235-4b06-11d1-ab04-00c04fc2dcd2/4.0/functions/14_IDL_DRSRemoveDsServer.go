package functions

import (
	"fmt"

	drsuapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// iDL_DRSRemoveDsServerRequest carries the [in] parameters of IDL_DRSRemoveDsServer.
type iDL_DRSRemoveDsServerRequest struct {
	HDrs        structures.DRS_HANDLE
	DwInVersion ndr.DWORD
	PmsgIn      structures.DRS_MSG_RMSVRREQ
}

func (*iDL_DRSRemoveDsServerRequest) Opnum() uint16 { return drsuapi.OpnumIDL_DRSRemoveDsServer }

// iDL_DRSRemoveDsServerResponse carries the [out] parameters and return value of IDL_DRSRemoveDsServer.
type iDL_DRSRemoveDsServerResponse struct {
	PdwOutVersion ndr.DWORD
	PmsgOut       structures.DRS_MSG_RMSVRREPLY
	Status        ndr.DWORD `ndr:"retval"`
}

// IDL_DRSRemoveDsServer calls IDL_DRSRemoveDsServer (opnum 14) ([MS-DRSR] — verify the parameter
// modeling and status handling).
func IDL_DRSRemoveDsServer(rpc ndr.Invoker, hDrs structures.DRS_HANDLE, dwInVersion ndr.DWORD, pmsgIn structures.DRS_MSG_RMSVRREQ) (PdwOutVersion ndr.DWORD, PmsgOut structures.DRS_MSG_RMSVRREPLY, err error) {
	req := &iDL_DRSRemoveDsServerRequest{
		HDrs:        hDrs,
		DwInVersion: dwInVersion,
		PmsgIn:      pmsgIn,
	}
	var resp iDL_DRSRemoveDsServerResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("IDL_DRSRemoveDsServer: %w", err)
		return
	}
	PdwOutVersion = resp.PdwOutVersion
	PmsgOut = resp.PmsgOut
	if uint32(resp.Status) != drsuapi.StatusSuccess {
		err = fmt.Errorf("IDL_DRSRemoveDsServer failed: %s", drsuapi.StatusString(uint32(resp.Status)))
	}
	return
}
