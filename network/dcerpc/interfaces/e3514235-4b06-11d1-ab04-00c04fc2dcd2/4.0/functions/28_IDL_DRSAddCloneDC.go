package functions

import (
	"fmt"

	drsuapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// iDL_DRSAddCloneDCRequest carries the [in] parameters of IDL_DRSAddCloneDC.
type iDL_DRSAddCloneDCRequest struct {
	HDrs        structures.DRS_HANDLE
	DwInVersion ndr.DWORD
	PmsgIn      structures.DRS_MSG_ADDCLONEDCREQ
}

func (*iDL_DRSAddCloneDCRequest) Opnum() uint16 { return drsuapi.OpnumIDL_DRSAddCloneDC }

// iDL_DRSAddCloneDCResponse carries the [out] parameters and return value of IDL_DRSAddCloneDC.
type iDL_DRSAddCloneDCResponse struct {
	PdwOutVersion ndr.DWORD
	PmsgOut       structures.DRS_MSG_ADDCLONEDCREPLY
	Status        ndr.DWORD `ndr:"retval"`
}

// IDL_DRSAddCloneDC calls IDL_DRSAddCloneDC (opnum 28) ([MS-DRSR] — verify the parameter
// modeling and status handling).
func IDL_DRSAddCloneDC(rpc ndr.Invoker, hDrs structures.DRS_HANDLE, dwInVersion ndr.DWORD, pmsgIn structures.DRS_MSG_ADDCLONEDCREQ) (PdwOutVersion ndr.DWORD, PmsgOut structures.DRS_MSG_ADDCLONEDCREPLY, err error) {
	req := &iDL_DRSAddCloneDCRequest{
		HDrs:        hDrs,
		DwInVersion: dwInVersion,
		PmsgIn:      pmsgIn,
	}
	var resp iDL_DRSAddCloneDCResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("IDL_DRSAddCloneDC: %w", err)
		return
	}
	PdwOutVersion = resp.PdwOutVersion
	PmsgOut = resp.PmsgOut
	if uint32(resp.Status) != drsuapi.StatusSuccess {
		err = fmt.Errorf("IDL_DRSAddCloneDC failed: %s", drsuapi.StatusString(uint32(resp.Status)))
	}
	return
}
