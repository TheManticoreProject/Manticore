package functions

import (
	"fmt"

	drsuapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// iDL_DRSFinishDemotionRequest carries the [in] parameters of IDL_DRSFinishDemotion.
type iDL_DRSFinishDemotionRequest struct {
	HDrs        structures.DRS_HANDLE
	DwInVersion ndr.DWORD
	PmsgIn      structures.DRS_MSG_FINISH_DEMOTIONREQ
}

func (*iDL_DRSFinishDemotionRequest) Opnum() uint16 { return drsuapi.OpnumIDL_DRSFinishDemotion }

// iDL_DRSFinishDemotionResponse carries the [out] parameters and return value of IDL_DRSFinishDemotion.
type iDL_DRSFinishDemotionResponse struct {
	PdwOutVersion ndr.DWORD
	PmsgOut       structures.DRS_MSG_FINISH_DEMOTIONREPLY
	Status        ndr.DWORD `ndr:"retval"`
}

// IDL_DRSFinishDemotion calls IDL_DRSFinishDemotion (opnum 27) ([MS-DRSR] — verify the parameter
// modeling and status handling).
func IDL_DRSFinishDemotion(rpc ndr.Invoker, hDrs structures.DRS_HANDLE, dwInVersion ndr.DWORD, pmsgIn structures.DRS_MSG_FINISH_DEMOTIONREQ) (PdwOutVersion ndr.DWORD, PmsgOut structures.DRS_MSG_FINISH_DEMOTIONREPLY, err error) {
	req := &iDL_DRSFinishDemotionRequest{
		HDrs:        hDrs,
		DwInVersion: dwInVersion,
		PmsgIn:      pmsgIn,
	}
	var resp iDL_DRSFinishDemotionResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("IDL_DRSFinishDemotion: %w", err)
		return
	}
	PdwOutVersion = resp.PdwOutVersion
	PmsgOut = resp.PmsgOut
	if uint32(resp.Status) != drsuapi.StatusSuccess {
		err = fmt.Errorf("IDL_DRSFinishDemotion failed: %s", drsuapi.StatusString(uint32(resp.Status)))
	}
	return
}
