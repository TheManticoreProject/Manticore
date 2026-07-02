package functions

import (
	"fmt"

	drsuapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdrsr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-drsr"
)

// iDL_DRSReplicaVerifyObjectsRequest carries the [in] parameters of IDL_DRSReplicaVerifyObjects.
type iDL_DRSReplicaVerifyObjectsRequest struct {
	HDrs       msdrsr.DRS_HANDLE
	DwVersion  ndr.DWORD
	PmsgVerify msdrsr.DRS_MSG_REPVERIFYOBJ
}

func (*iDL_DRSReplicaVerifyObjectsRequest) Opnum() uint16 {
	return drsuapi.OpnumIDL_DRSReplicaVerifyObjects
}

// iDL_DRSReplicaVerifyObjectsResponse carries the [out] parameters and return value of IDL_DRSReplicaVerifyObjects.
type iDL_DRSReplicaVerifyObjectsResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// IDL_DRSReplicaVerifyObjects calls IDL_DRSReplicaVerifyObjects (opnum 22) ([MS-DRSR] — verify the parameter
// modeling and status handling).
func IDL_DRSReplicaVerifyObjects(rpc ndr.Invoker, hDrs msdrsr.DRS_HANDLE, dwVersion ndr.DWORD, pmsgVerify msdrsr.DRS_MSG_REPVERIFYOBJ) (err error) {
	req := &iDL_DRSReplicaVerifyObjectsRequest{
		HDrs:       hDrs,
		DwVersion:  dwVersion,
		PmsgVerify: pmsgVerify,
	}
	var resp iDL_DRSReplicaVerifyObjectsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("IDL_DRSReplicaVerifyObjects: %w", err)
		return
	}
	if uint32(resp.Status) != drsuapi.StatusSuccess {
		err = fmt.Errorf("IDL_DRSReplicaVerifyObjects failed: %s", drsuapi.StatusString(uint32(resp.Status)))
	}
	return
}
