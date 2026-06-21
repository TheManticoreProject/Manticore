package functions

import (
	"fmt"

	drsuapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// iDL_DRSReplicaDelRequest carries the [in] parameters of IDL_DRSReplicaDel.
type iDL_DRSReplicaDelRequest struct {
	HDrs      structures.DRS_HANDLE
	DwVersion ndr.DWORD
	PmsgDel   structures.DRS_MSG_REPDEL
}

func (*iDL_DRSReplicaDelRequest) Opnum() uint16 { return drsuapi.OpnumIDL_DRSReplicaDel }

// iDL_DRSReplicaDelResponse carries the [out] parameters and return value of IDL_DRSReplicaDel.
type iDL_DRSReplicaDelResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// IDL_DRSReplicaDel calls IDL_DRSReplicaDel (opnum 6) ([MS-DRSR] — verify the parameter
// modeling and status handling).
func IDL_DRSReplicaDel(rpc ndr.Invoker, hDrs structures.DRS_HANDLE, dwVersion ndr.DWORD, pmsgDel structures.DRS_MSG_REPDEL) (err error) {
	req := &iDL_DRSReplicaDelRequest{
		HDrs:      hDrs,
		DwVersion: dwVersion,
		PmsgDel:   pmsgDel,
	}
	var resp iDL_DRSReplicaDelResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("IDL_DRSReplicaDel: %w", err)
		return
	}
	if uint32(resp.Status) != drsuapi.StatusSuccess {
		err = fmt.Errorf("IDL_DRSReplicaDel failed: %s", drsuapi.StatusString(uint32(resp.Status)))
	}
	return
}
