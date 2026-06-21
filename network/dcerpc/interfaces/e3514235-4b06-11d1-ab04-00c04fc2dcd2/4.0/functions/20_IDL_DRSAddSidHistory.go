package functions

import (
	"fmt"

	drsuapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// iDL_DRSAddSidHistoryRequest carries the [in] parameters of IDL_DRSAddSidHistory.
type iDL_DRSAddSidHistoryRequest struct {
	HDrs        structures.DRS_HANDLE
	DwInVersion ndr.DWORD
	PmsgIn      structures.DRS_MSG_ADDSIDREQ
}

func (*iDL_DRSAddSidHistoryRequest) Opnum() uint16 { return drsuapi.OpnumIDL_DRSAddSidHistory }

// iDL_DRSAddSidHistoryResponse carries the [out] parameters and return value of IDL_DRSAddSidHistory.
type iDL_DRSAddSidHistoryResponse struct {
	PdwOutVersion ndr.DWORD
	PmsgOut       structures.DRS_MSG_ADDSIDREPLY
	Status        ndr.DWORD `ndr:"retval"`
}

// IDL_DRSAddSidHistory calls IDL_DRSAddSidHistory (opnum 20) ([MS-DRSR] — verify the parameter
// modeling and status handling).
func IDL_DRSAddSidHistory(rpc ndr.Invoker, hDrs structures.DRS_HANDLE, dwInVersion ndr.DWORD, pmsgIn structures.DRS_MSG_ADDSIDREQ) (PdwOutVersion ndr.DWORD, PmsgOut structures.DRS_MSG_ADDSIDREPLY, err error) {
	req := &iDL_DRSAddSidHistoryRequest{
		HDrs:        hDrs,
		DwInVersion: dwInVersion,
		PmsgIn:      pmsgIn,
	}
	var resp iDL_DRSAddSidHistoryResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("IDL_DRSAddSidHistory: %w", err)
		return
	}
	PdwOutVersion = resp.PdwOutVersion
	PmsgOut = resp.PmsgOut
	if uint32(resp.Status) != drsuapi.StatusSuccess {
		err = fmt.Errorf("IDL_DRSAddSidHistory failed: %s", drsuapi.StatusString(uint32(resp.Status)))
	}
	return
}
