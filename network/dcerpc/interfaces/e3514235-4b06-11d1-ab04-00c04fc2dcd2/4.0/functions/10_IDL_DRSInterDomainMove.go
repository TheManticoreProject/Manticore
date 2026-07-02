package functions

import (
	"fmt"

	drsuapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdrsr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-drsr"
)

// iDL_DRSInterDomainMoveRequest carries the [in] parameters of IDL_DRSInterDomainMove.
type iDL_DRSInterDomainMoveRequest struct {
	HDrs        msdrsr.DRS_HANDLE
	DwInVersion ndr.DWORD
	PmsgIn      msdrsr.DRS_MSG_MOVEREQ
}

func (*iDL_DRSInterDomainMoveRequest) Opnum() uint16 { return drsuapi.OpnumIDL_DRSInterDomainMove }

// iDL_DRSInterDomainMoveResponse carries the [out] parameters and return value of IDL_DRSInterDomainMove.
type iDL_DRSInterDomainMoveResponse struct {
	PdwOutVersion ndr.DWORD
	PmsgOut       msdrsr.DRS_MSG_MOVEREPLY
	Status        ndr.DWORD `ndr:"retval"`
}

// IDL_DRSInterDomainMove calls IDL_DRSInterDomainMove (opnum 10) ([MS-DRSR] — verify the parameter
// modeling and status handling).
func IDL_DRSInterDomainMove(rpc ndr.Invoker, hDrs msdrsr.DRS_HANDLE, dwInVersion ndr.DWORD, pmsgIn msdrsr.DRS_MSG_MOVEREQ) (PdwOutVersion ndr.DWORD, PmsgOut msdrsr.DRS_MSG_MOVEREPLY, err error) {
	req := &iDL_DRSInterDomainMoveRequest{
		HDrs:        hDrs,
		DwInVersion: dwInVersion,
		PmsgIn:      pmsgIn,
	}
	var resp iDL_DRSInterDomainMoveResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("IDL_DRSInterDomainMove: %w", err)
		return
	}
	PdwOutVersion = resp.PdwOutVersion
	PmsgOut = resp.PmsgOut
	if uint32(resp.Status) != drsuapi.StatusSuccess {
		err = fmt.Errorf("IDL_DRSInterDomainMove failed: %s", drsuapi.StatusString(uint32(resp.Status)))
	}
	return
}
