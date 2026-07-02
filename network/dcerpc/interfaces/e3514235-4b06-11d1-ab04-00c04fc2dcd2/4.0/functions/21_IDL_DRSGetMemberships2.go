package functions

import (
	"fmt"

	drsuapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdrsr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-drsr"
)

// iDL_DRSGetMemberships2Request carries the [in] parameters of IDL_DRSGetMemberships2.
type iDL_DRSGetMemberships2Request struct {
	HDrs        msdrsr.DRS_HANDLE
	DwInVersion ndr.DWORD
	PmsgIn      msdrsr.DRS_MSG_GETMEMBERSHIPS2_REQ
}

func (*iDL_DRSGetMemberships2Request) Opnum() uint16 { return drsuapi.OpnumIDL_DRSGetMemberships2 }

// iDL_DRSGetMemberships2Response carries the [out] parameters and return value of IDL_DRSGetMemberships2.
type iDL_DRSGetMemberships2Response struct {
	PdwOutVersion ndr.DWORD
	PmsgOut       msdrsr.DRS_MSG_GETMEMBERSHIPS2_REPLY
	Status        ndr.DWORD `ndr:"retval"`
}

// IDL_DRSGetMemberships2 calls IDL_DRSGetMemberships2 (opnum 21) ([MS-DRSR] — verify the parameter
// modeling and status handling).
func IDL_DRSGetMemberships2(rpc ndr.Invoker, hDrs msdrsr.DRS_HANDLE, dwInVersion ndr.DWORD, pmsgIn msdrsr.DRS_MSG_GETMEMBERSHIPS2_REQ) (PdwOutVersion ndr.DWORD, PmsgOut msdrsr.DRS_MSG_GETMEMBERSHIPS2_REPLY, err error) {
	req := &iDL_DRSGetMemberships2Request{
		HDrs:        hDrs,
		DwInVersion: dwInVersion,
		PmsgIn:      pmsgIn,
	}
	var resp iDL_DRSGetMemberships2Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("IDL_DRSGetMemberships2: %w", err)
		return
	}
	PdwOutVersion = resp.PdwOutVersion
	PmsgOut = resp.PmsgOut
	if uint32(resp.Status) != drsuapi.StatusSuccess {
		err = fmt.Errorf("IDL_DRSGetMemberships2 failed: %s", drsuapi.StatusString(uint32(resp.Status)))
	}
	return
}
