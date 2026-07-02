package functions

import (
	"fmt"

	drsuapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdrsr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-drsr"
)

// iDL_DRSReplicaDemotionRequest carries the [in] parameters of IDL_DRSReplicaDemotion.
type iDL_DRSReplicaDemotionRequest struct {
	HDrs        msdrsr.DRS_HANDLE
	DwInVersion ndr.DWORD
	PmsgIn      msdrsr.DRS_MSG_REPLICA_DEMOTIONREQ
}

func (*iDL_DRSReplicaDemotionRequest) Opnum() uint16 { return drsuapi.OpnumIDL_DRSReplicaDemotion }

// iDL_DRSReplicaDemotionResponse carries the [out] parameters and return value of IDL_DRSReplicaDemotion.
type iDL_DRSReplicaDemotionResponse struct {
	PdwOutVersion ndr.DWORD
	PmsgOut       msdrsr.DRS_MSG_REPLICA_DEMOTIONREPLY
	Status        ndr.DWORD `ndr:"retval"`
}

// IDL_DRSReplicaDemotion calls IDL_DRSReplicaDemotion (opnum 26) ([MS-DRSR] — verify the parameter
// modeling and status handling).
func IDL_DRSReplicaDemotion(rpc ndr.Invoker, hDrs msdrsr.DRS_HANDLE, dwInVersion ndr.DWORD, pmsgIn msdrsr.DRS_MSG_REPLICA_DEMOTIONREQ) (PdwOutVersion ndr.DWORD, PmsgOut msdrsr.DRS_MSG_REPLICA_DEMOTIONREPLY, err error) {
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
