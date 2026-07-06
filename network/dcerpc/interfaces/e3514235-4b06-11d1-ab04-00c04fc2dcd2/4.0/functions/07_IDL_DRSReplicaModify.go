package functions

// IDL source: [MS-DRSR] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-drsr/3f5d9495-9563-44de-876a-ce6f880e3fb2
// A fetched copy is kept at ms-drsr.idl in the interface directory.

import (
	"fmt"

	drsuapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdrsr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-drsr"
)

// iDL_DRSReplicaModifyRequest carries the [in] parameters of IDL_DRSReplicaModify.
type iDL_DRSReplicaModifyRequest struct {
	HDrs      msdrsr.DRS_HANDLE
	DwVersion ndr.DWORD
	PmsgMod   msdrsr.DRS_MSG_REPMOD
}

func (*iDL_DRSReplicaModifyRequest) Opnum() uint16 { return drsuapi.OpnumIDL_DRSReplicaModify }

// iDL_DRSReplicaModifyResponse carries the [out] parameters and return value of IDL_DRSReplicaModify.
type iDL_DRSReplicaModifyResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// IDL_DRSReplicaModify calls IDL_DRSReplicaModify (opnum 7) ([MS-DRSR] — verify the parameter
// modeling and status handling).
func IDL_DRSReplicaModify(rpc ndr.Invoker, hDrs msdrsr.DRS_HANDLE, dwVersion ndr.DWORD, pmsgMod msdrsr.DRS_MSG_REPMOD) (err error) {
	req := &iDL_DRSReplicaModifyRequest{
		HDrs:      hDrs,
		DwVersion: dwVersion,
		PmsgMod:   pmsgMod,
	}
	var resp iDL_DRSReplicaModifyResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("IDL_DRSReplicaModify: %w", err)
		return
	}
	if uint32(resp.Status) != drsuapi.StatusSuccess {
		err = fmt.Errorf("IDL_DRSReplicaModify failed: %s", drsuapi.StatusString(uint32(resp.Status)))
	}
	return
}
