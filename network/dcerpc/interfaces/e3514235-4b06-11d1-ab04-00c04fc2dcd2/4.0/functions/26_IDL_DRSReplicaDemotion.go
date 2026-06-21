package functions

import (
	"fmt"

	drsuapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// iDL_DRSReplicaDemotionRequest carries the [in] parameters of IDL_DRSReplicaDemotion.
type iDL_DRSReplicaDemotionRequest struct {
	HDrs        structures.DRS_HANDLE
	DwInVersion ndr.DWORD
	PmsgIn      structures.DRS_MSG_REPLICA_DEMOTIONREQ
}

func (*iDL_DRSReplicaDemotionRequest) Opnum() uint16 { return drsuapi.OpnumIDL_DRSReplicaDemotion }

// iDL_DRSReplicaDemotionResponse carries the [out] parameters and return value of IDL_DRSReplicaDemotion.
type iDL_DRSReplicaDemotionResponse struct {
	PdwOutVersion ndr.DWORD
	PmsgOut       structures.DRS_MSG_REPLICA_DEMOTIONREPLY
	Status        ndr.DWORD `ndr:"retval"`
}

// IDL_DRSReplicaDemotion calls IDL_DRSReplicaDemotion (opnum 26) ([MS-DRSR] — verify the parameter
// modeling and status handling).
func IDL_DRSReplicaDemotion(rpc ndr.Invoker, hDrs structures.DRS_HANDLE, dwInVersion ndr.DWORD, pmsgIn structures.DRS_MSG_REPLICA_DEMOTIONREQ) (PdwOutVersion ndr.DWORD, PmsgOut structures.DRS_MSG_REPLICA_DEMOTIONREPLY, err error) {
	req := &iDL_DRSReplicaDemotionRequest{
		HDrs:        hDrs,
		DwInVersion: dwInVersion,
		PmsgIn:      pmsgIn,
	}
	var resp iDL_DRSReplicaDemotionResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("IDL_DRSReplicaDemotion: %w", err)
		return
	}
	PdwOutVersion = resp.PdwOutVersion
	PmsgOut = resp.PmsgOut
	if uint32(resp.Status) != drsuapi.StatusSuccess {
		err = fmt.Errorf("IDL_DRSReplicaDemotion failed: %s", drsuapi.StatusString(uint32(resp.Status)))
	}
	return
}
