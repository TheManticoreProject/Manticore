package functions

import (
	"fmt"

	drsuapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdrsr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-drsr"
)

// iDL_DRSWriteNgcKeyRequest carries the [in] parameters of IDL_DRSWriteNgcKey.
type iDL_DRSWriteNgcKeyRequest struct {
	HDrs        msdrsr.DRS_HANDLE
	DwInVersion ndr.DWORD
	PmsgIn      msdrsr.DRS_MSG_WRITENGCKEYREQ
}

func (*iDL_DRSWriteNgcKeyRequest) Opnum() uint16 { return drsuapi.OpnumIDL_DRSWriteNgcKey }

// iDL_DRSWriteNgcKeyResponse carries the [out] parameters and return value of IDL_DRSWriteNgcKey.
type iDL_DRSWriteNgcKeyResponse struct {
	PdwOutVersion ndr.DWORD
	PmsgOut       msdrsr.DRS_MSG_WRITENGCKEYREPLY
	Status        ndr.DWORD `ndr:"retval"`
}

// IDL_DRSWriteNgcKey calls IDL_DRSWriteNgcKey (opnum 29) ([MS-DRSR] — verify the parameter
// modeling and status handling).
func IDL_DRSWriteNgcKey(rpc ndr.Invoker, hDrs msdrsr.DRS_HANDLE, dwInVersion ndr.DWORD, pmsgIn msdrsr.DRS_MSG_WRITENGCKEYREQ) (PdwOutVersion ndr.DWORD, PmsgOut msdrsr.DRS_MSG_WRITENGCKEYREPLY, err error) {
	req := &iDL_DRSWriteNgcKeyRequest{
		HDrs:        hDrs,
		DwInVersion: dwInVersion,
		PmsgIn:      pmsgIn,
	}
	var resp iDL_DRSWriteNgcKeyResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("IDL_DRSWriteNgcKey: %w", err)
		return
	}
	PdwOutVersion = resp.PdwOutVersion
	PmsgOut = resp.PmsgOut
	if uint32(resp.Status) != drsuapi.StatusSuccess {
		err = fmt.Errorf("IDL_DRSWriteNgcKey failed: %s", drsuapi.StatusString(uint32(resp.Status)))
	}
	return
}
