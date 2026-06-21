package functions

import (
	"fmt"

	drsuapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// iDL_DRSInitDemotionRequest carries the [in] parameters of IDL_DRSInitDemotion.
type iDL_DRSInitDemotionRequest struct {
	HDrs        structures.DRS_HANDLE
	DwInVersion ndr.DWORD
	PmsgIn      structures.DRS_MSG_INIT_DEMOTIONREQ
}

func (*iDL_DRSInitDemotionRequest) Opnum() uint16 { return drsuapi.OpnumIDL_DRSInitDemotion }

// iDL_DRSInitDemotionResponse carries the [out] parameters and return value of IDL_DRSInitDemotion.
type iDL_DRSInitDemotionResponse struct {
	PdwOutVersion ndr.DWORD
	PmsgOut       structures.DRS_MSG_INIT_DEMOTIONREPLY
	Status        ndr.DWORD `ndr:"retval"`
}

// IDL_DRSInitDemotion calls IDL_DRSInitDemotion (opnum 25) ([MS-DRSR] — verify the parameter
// modeling and status handling).
func IDL_DRSInitDemotion(rpc ndr.Invoker, hDrs structures.DRS_HANDLE, dwInVersion ndr.DWORD, pmsgIn structures.DRS_MSG_INIT_DEMOTIONREQ) (PdwOutVersion ndr.DWORD, PmsgOut structures.DRS_MSG_INIT_DEMOTIONREPLY, err error) {
	req := &iDL_DRSInitDemotionRequest{
		HDrs:        hDrs,
		DwInVersion: dwInVersion,
		PmsgIn:      pmsgIn,
	}
	var resp iDL_DRSInitDemotionResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("IDL_DRSInitDemotion: %w", err)
		return
	}
	PdwOutVersion = resp.PdwOutVersion
	PmsgOut = resp.PmsgOut
	if uint32(resp.Status) != drsuapi.StatusSuccess {
		err = fmt.Errorf("IDL_DRSInitDemotion failed: %s", drsuapi.StatusString(uint32(resp.Status)))
	}
	return
}
