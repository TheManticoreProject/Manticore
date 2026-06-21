package functions

import (
	"fmt"

	drsuapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// iDL_DRSReplicaAddRequest carries the [in] parameters of IDL_DRSReplicaAdd.
type iDL_DRSReplicaAddRequest struct {
	HDrs      structures.DRS_HANDLE
	DwVersion ndr.DWORD
	PmsgAdd   structures.DRS_MSG_REPADD
}

func (*iDL_DRSReplicaAddRequest) Opnum() uint16 { return drsuapi.OpnumIDL_DRSReplicaAdd }

// iDL_DRSReplicaAddResponse carries the [out] parameters and return value of IDL_DRSReplicaAdd.
type iDL_DRSReplicaAddResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// IDL_DRSReplicaAdd calls IDL_DRSReplicaAdd (opnum 5) ([MS-DRSR] — verify the parameter
// modeling and status handling).
func IDL_DRSReplicaAdd(rpc ndr.Invoker, hDrs structures.DRS_HANDLE, dwVersion ndr.DWORD, pmsgAdd structures.DRS_MSG_REPADD) (err error) {
	req := &iDL_DRSReplicaAddRequest{
		HDrs:      hDrs,
		DwVersion: dwVersion,
		PmsgAdd:   pmsgAdd,
	}
	var resp iDL_DRSReplicaAddResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("IDL_DRSReplicaAdd: %w", err)
		return
	}
	if uint32(resp.Status) != drsuapi.StatusSuccess {
		err = fmt.Errorf("IDL_DRSReplicaAdd failed: %s", drsuapi.StatusString(uint32(resp.Status)))
	}
	return
}
