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

// iDL_DRSReplicaDelRequest carries the [in] parameters of IDL_DRSReplicaDel.
type iDL_DRSReplicaDelRequest struct {
	HDrs      msdrsr.DRS_HANDLE
	DwVersion ndr.DWORD
	PmsgDel   msdrsr.DRS_MSG_REPDEL
}

func (*iDL_DRSReplicaDelRequest) Opnum() uint16 { return drsuapi.OpnumIDL_DRSReplicaDel }

// iDL_DRSReplicaDelResponse carries the [out] parameters and return value of IDL_DRSReplicaDel.
type iDL_DRSReplicaDelResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// IDL_DRSReplicaDel calls IDL_DRSReplicaDel (opnum 6) ([MS-DRSR] — verify the parameter
// modeling and status handling).
func IDL_DRSReplicaDel(rpc ndr.Invoker, hDrs msdrsr.DRS_HANDLE, dwVersion ndr.DWORD, pmsgDel msdrsr.DRS_MSG_REPDEL) (err error) {
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
