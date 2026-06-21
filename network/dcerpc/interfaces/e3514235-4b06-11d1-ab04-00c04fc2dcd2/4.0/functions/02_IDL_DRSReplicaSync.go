package functions

import (
	"fmt"

	drsuapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// iDL_DRSReplicaSyncRequest carries the [in] parameters of IDL_DRSReplicaSync.
type iDL_DRSReplicaSyncRequest struct {
	HDrs      structures.DRS_HANDLE
	DwVersion ndr.DWORD
	PmsgSync  structures.DRS_MSG_REPSYNC
}

func (*iDL_DRSReplicaSyncRequest) Opnum() uint16 { return drsuapi.OpnumIDL_DRSReplicaSync }

// iDL_DRSReplicaSyncResponse carries the [out] parameters and return value of IDL_DRSReplicaSync.
type iDL_DRSReplicaSyncResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// IDL_DRSReplicaSync calls IDL_DRSReplicaSync (opnum 2) ([MS-DRSR] — verify the parameter
// modeling and status handling).
func IDL_DRSReplicaSync(rpc ndr.Invoker, hDrs structures.DRS_HANDLE, dwVersion ndr.DWORD, pmsgSync structures.DRS_MSG_REPSYNC) (err error) {
	req := &iDL_DRSReplicaSyncRequest{
		HDrs:      hDrs,
		DwVersion: dwVersion,
		PmsgSync:  pmsgSync,
	}
	var resp iDL_DRSReplicaSyncResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("IDL_DRSReplicaSync: %w", err)
		return
	}
	if uint32(resp.Status) != drsuapi.StatusSuccess {
		err = fmt.Errorf("IDL_DRSReplicaSync failed: %s", drsuapi.StatusString(uint32(resp.Status)))
	}
	return
}
