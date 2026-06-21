package functions

import (
	"fmt"

	drsuapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// iDL_DRSAddEntryRequest carries the [in] parameters of IDL_DRSAddEntry.
type iDL_DRSAddEntryRequest struct {
	HDrs        structures.DRS_HANDLE
	DwInVersion ndr.DWORD
	PmsgIn      structures.DRS_MSG_ADDENTRYREQ
}

func (*iDL_DRSAddEntryRequest) Opnum() uint16 { return drsuapi.OpnumIDL_DRSAddEntry }

// iDL_DRSAddEntryResponse carries the [out] parameters and return value of IDL_DRSAddEntry.
type iDL_DRSAddEntryResponse struct {
	PdwOutVersion ndr.DWORD
	PmsgOut       structures.DRS_MSG_ADDENTRYREPLY
	Status        ndr.DWORD `ndr:"retval"`
}

// IDL_DRSAddEntry calls IDL_DRSAddEntry (opnum 17) ([MS-DRSR] — verify the parameter
// modeling and status handling).
func IDL_DRSAddEntry(rpc ndr.Invoker, hDrs structures.DRS_HANDLE, dwInVersion ndr.DWORD, pmsgIn structures.DRS_MSG_ADDENTRYREQ) (PdwOutVersion ndr.DWORD, PmsgOut structures.DRS_MSG_ADDENTRYREPLY, err error) {
	req := &iDL_DRSAddEntryRequest{
		HDrs:        hDrs,
		DwInVersion: dwInVersion,
		PmsgIn:      pmsgIn,
	}
	var resp iDL_DRSAddEntryResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("IDL_DRSAddEntry: %w", err)
		return
	}
	PdwOutVersion = resp.PdwOutVersion
	PmsgOut = resp.PmsgOut
	if uint32(resp.Status) != drsuapi.StatusSuccess {
		err = fmt.Errorf("IDL_DRSAddEntry failed: %s", drsuapi.StatusString(uint32(resp.Status)))
	}
	return
}
