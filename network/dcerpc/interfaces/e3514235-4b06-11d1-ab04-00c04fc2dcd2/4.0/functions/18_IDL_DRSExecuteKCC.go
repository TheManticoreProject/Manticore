package functions

import (
	"fmt"

	drsuapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// iDL_DRSExecuteKCCRequest carries the [in] parameters of IDL_DRSExecuteKCC.
type iDL_DRSExecuteKCCRequest struct {
	HDrs        structures.DRS_HANDLE
	DwInVersion ndr.DWORD
	PmsgIn      structures.DRS_MSG_KCC_EXECUTE
}

func (*iDL_DRSExecuteKCCRequest) Opnum() uint16 { return drsuapi.OpnumIDL_DRSExecuteKCC }

// iDL_DRSExecuteKCCResponse carries the [out] parameters and return value of IDL_DRSExecuteKCC.
type iDL_DRSExecuteKCCResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// IDL_DRSExecuteKCC calls IDL_DRSExecuteKCC (opnum 18) ([MS-DRSR] — verify the parameter
// modeling and status handling).
func IDL_DRSExecuteKCC(rpc ndr.Invoker, hDrs structures.DRS_HANDLE, dwInVersion ndr.DWORD, pmsgIn structures.DRS_MSG_KCC_EXECUTE) (err error) {
	req := &iDL_DRSExecuteKCCRequest{
		HDrs:        hDrs,
		DwInVersion: dwInVersion,
		PmsgIn:      pmsgIn,
	}
	var resp iDL_DRSExecuteKCCResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("IDL_DRSExecuteKCC: %w", err)
		return
	}
	if uint32(resp.Status) != drsuapi.StatusSuccess {
		err = fmt.Errorf("IDL_DRSExecuteKCC failed: %s", drsuapi.StatusString(uint32(resp.Status)))
	}
	return
}
