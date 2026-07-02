package functions

import (
	"fmt"

	drsuapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdrsr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-drsr"
)

// iDL_DRSReadNgcKeyRequest carries the [in] parameters of IDL_DRSReadNgcKey.
type iDL_DRSReadNgcKeyRequest struct {
	HDrs        msdrsr.DRS_HANDLE
	DwInVersion ndr.DWORD
	PmsgIn      msdrsr.DRS_MSG_READNGCKEYREQ
}

func (*iDL_DRSReadNgcKeyRequest) Opnum() uint16 { return drsuapi.OpnumIDL_DRSReadNgcKey }

// iDL_DRSReadNgcKeyResponse carries the [out] parameters and return value of IDL_DRSReadNgcKey.
type iDL_DRSReadNgcKeyResponse struct {
	PdwOutVersion ndr.DWORD
	PmsgOut       msdrsr.DRS_MSG_READNGCKEYREPLY
	Status        ndr.DWORD `ndr:"retval"`
}

// IDL_DRSReadNgcKey calls IDL_DRSReadNgcKey (opnum 30) ([MS-DRSR] — verify the parameter
// modeling and status handling).
func IDL_DRSReadNgcKey(rpc ndr.Invoker, hDrs msdrsr.DRS_HANDLE, dwInVersion ndr.DWORD, pmsgIn msdrsr.DRS_MSG_READNGCKEYREQ) (PdwOutVersion ndr.DWORD, PmsgOut msdrsr.DRS_MSG_READNGCKEYREPLY, err error) {
	req := &iDL_DRSReadNgcKeyRequest{
		HDrs:        hDrs,
		DwInVersion: dwInVersion,
		PmsgIn:      pmsgIn,
	}
	var resp iDL_DRSReadNgcKeyResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("IDL_DRSReadNgcKey: %w", err)
		return
	}
	PdwOutVersion = resp.PdwOutVersion
	PmsgOut = resp.PmsgOut
	if uint32(resp.Status) != drsuapi.StatusSuccess {
		err = fmt.Errorf("IDL_DRSReadNgcKey failed: %s", drsuapi.StatusString(uint32(resp.Status)))
	}
	return
}
