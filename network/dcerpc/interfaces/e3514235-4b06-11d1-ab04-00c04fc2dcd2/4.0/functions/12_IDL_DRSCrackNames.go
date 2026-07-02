package functions

import (
	"fmt"

	drsuapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdrsr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-drsr"
)

// iDL_DRSCrackNamesRequest carries the [in] parameters of IDL_DRSCrackNames.
type iDL_DRSCrackNamesRequest struct {
	HDrs        msdrsr.DRS_HANDLE
	DwInVersion ndr.DWORD
	PmsgIn      msdrsr.DRS_MSG_CRACKREQ
}

func (*iDL_DRSCrackNamesRequest) Opnum() uint16 { return drsuapi.OpnumIDL_DRSCrackNames }

// iDL_DRSCrackNamesResponse carries the [out] parameters and return value of IDL_DRSCrackNames.
type iDL_DRSCrackNamesResponse struct {
	PdwOutVersion ndr.DWORD
	PmsgOut       msdrsr.DRS_MSG_CRACKREPLY
	Status        ndr.DWORD `ndr:"retval"`
}

// IDL_DRSCrackNames calls IDL_DRSCrackNames (opnum 12) ([MS-DRSR] — verify the parameter
// modeling and status handling).
func IDL_DRSCrackNames(rpc ndr.Invoker, hDrs msdrsr.DRS_HANDLE, dwInVersion ndr.DWORD, pmsgIn msdrsr.DRS_MSG_CRACKREQ) (PdwOutVersion ndr.DWORD, PmsgOut msdrsr.DRS_MSG_CRACKREPLY, err error) {
	req := &iDL_DRSCrackNamesRequest{
		HDrs:        hDrs,
		DwInVersion: dwInVersion,
		PmsgIn:      pmsgIn,
	}
	var resp iDL_DRSCrackNamesResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("IDL_DRSCrackNames: %w", err)
		return
	}
	PdwOutVersion = resp.PdwOutVersion
	PmsgOut = resp.PmsgOut
	if uint32(resp.Status) != drsuapi.StatusSuccess {
		err = fmt.Errorf("IDL_DRSCrackNames failed: %s", drsuapi.StatusString(uint32(resp.Status)))
	}
	return
}
