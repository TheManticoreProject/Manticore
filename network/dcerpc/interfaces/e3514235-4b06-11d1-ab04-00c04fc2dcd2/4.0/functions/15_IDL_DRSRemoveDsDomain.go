package functions

import (
	"fmt"

	drsuapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// iDL_DRSRemoveDsDomainRequest carries the [in] parameters of IDL_DRSRemoveDsDomain.
type iDL_DRSRemoveDsDomainRequest struct {
	HDrs        structures.DRS_HANDLE
	DwInVersion ndr.DWORD
	PmsgIn      structures.DRS_MSG_RMDMNREQ
}

func (*iDL_DRSRemoveDsDomainRequest) Opnum() uint16 { return drsuapi.OpnumIDL_DRSRemoveDsDomain }

// iDL_DRSRemoveDsDomainResponse carries the [out] parameters and return value of IDL_DRSRemoveDsDomain.
type iDL_DRSRemoveDsDomainResponse struct {
	PdwOutVersion ndr.DWORD
	PmsgOut       structures.DRS_MSG_RMDMNREPLY
	Status        ndr.DWORD `ndr:"retval"`
}

// IDL_DRSRemoveDsDomain calls IDL_DRSRemoveDsDomain (opnum 15) ([MS-DRSR] — verify the parameter
// modeling and status handling).
func IDL_DRSRemoveDsDomain(rpc ndr.Invoker, hDrs structures.DRS_HANDLE, dwInVersion ndr.DWORD, pmsgIn structures.DRS_MSG_RMDMNREQ) (PdwOutVersion ndr.DWORD, PmsgOut structures.DRS_MSG_RMDMNREPLY, err error) {
	req := &iDL_DRSRemoveDsDomainRequest{
		HDrs:        hDrs,
		DwInVersion: dwInVersion,
		PmsgIn:      pmsgIn,
	}
	var resp iDL_DRSRemoveDsDomainResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("IDL_DRSRemoveDsDomain: %w", err)
		return
	}
	PdwOutVersion = resp.PdwOutVersion
	PmsgOut = resp.PmsgOut
	if uint32(resp.Status) != drsuapi.StatusSuccess {
		err = fmt.Errorf("IDL_DRSRemoveDsDomain failed: %s", drsuapi.StatusString(uint32(resp.Status)))
	}
	return
}
