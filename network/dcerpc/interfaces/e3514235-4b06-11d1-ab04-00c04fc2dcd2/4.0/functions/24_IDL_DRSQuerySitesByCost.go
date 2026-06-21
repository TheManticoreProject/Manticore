package functions

import (
	"fmt"

	drsuapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// iDL_DRSQuerySitesByCostRequest carries the [in] parameters of IDL_DRSQuerySitesByCost.
type iDL_DRSQuerySitesByCostRequest struct {
	HDrs        structures.DRS_HANDLE
	DwInVersion ndr.DWORD
	PmsgIn      structures.DRS_MSG_QUERYSITESREQ
}

func (*iDL_DRSQuerySitesByCostRequest) Opnum() uint16 { return drsuapi.OpnumIDL_DRSQuerySitesByCost }

// iDL_DRSQuerySitesByCostResponse carries the [out] parameters and return value of IDL_DRSQuerySitesByCost.
type iDL_DRSQuerySitesByCostResponse struct {
	PdwOutVersion ndr.DWORD
	PmsgOut       structures.DRS_MSG_QUERYSITESREPLY
	Status        ndr.DWORD `ndr:"retval"`
}

// IDL_DRSQuerySitesByCost calls IDL_DRSQuerySitesByCost (opnum 24) ([MS-DRSR] — verify the parameter
// modeling and status handling).
func IDL_DRSQuerySitesByCost(rpc ndr.Invoker, hDrs structures.DRS_HANDLE, dwInVersion ndr.DWORD, pmsgIn structures.DRS_MSG_QUERYSITESREQ) (PdwOutVersion ndr.DWORD, PmsgOut structures.DRS_MSG_QUERYSITESREPLY, err error) {
	req := &iDL_DRSQuerySitesByCostRequest{
		HDrs:        hDrs,
		DwInVersion: dwInVersion,
		PmsgIn:      pmsgIn,
	}
	var resp iDL_DRSQuerySitesByCostResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("IDL_DRSQuerySitesByCost: %w", err)
		return
	}
	PdwOutVersion = resp.PdwOutVersion
	PmsgOut = resp.PmsgOut
	if uint32(resp.Status) != drsuapi.StatusSuccess {
		err = fmt.Errorf("IDL_DRSQuerySitesByCost failed: %s", drsuapi.StatusString(uint32(resp.Status)))
	}
	return
}
