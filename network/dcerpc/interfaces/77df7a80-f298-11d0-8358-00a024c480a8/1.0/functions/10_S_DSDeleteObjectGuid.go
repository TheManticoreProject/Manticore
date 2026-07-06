package functions

import (
	"fmt"

	dscomm "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/77df7a80-f298-11d0-8358-00a024c480a8/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// s_DSDeleteObjectGuidRequest carries the [in] parameters of S_DSDeleteObjectGuid.
type s_DSDeleteObjectGuidRequest struct {
	DwObjectType ndr.DWORD
	PGuid        msdtyp.GUID
}

func (*s_DSDeleteObjectGuidRequest) Opnum() uint16 { return dscomm.OpnumS_DSDeleteObjectGuid }

// s_DSDeleteObjectGuidResponse carries the [out] parameters and return value of S_DSDeleteObjectGuid.
type s_DSDeleteObjectGuidResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// S_DSDeleteObjectGuid calls S_DSDeleteObjectGuid (opnum 10) ([MS-MQDS] — verify the parameter
// modeling and status handling).
func S_DSDeleteObjectGuid(rpc ndr.Invoker, dwObjectType ndr.DWORD, pGuid msdtyp.GUID) (err error) {
	req := &s_DSDeleteObjectGuidRequest{
		DwObjectType: dwObjectType,
		PGuid:        pGuid,
	}
	var resp s_DSDeleteObjectGuidResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("S_DSDeleteObjectGuid: %w", err)
		return
	}
	if uint32(resp.Status) != dscomm.StatusSuccess {
		err = fmt.Errorf("S_DSDeleteObjectGuid failed: %s", dscomm.StatusString(uint32(resp.Status)))
	}
	return
}
