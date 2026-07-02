package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	dscomm "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/77df7a80-f298-11d0-8358-00a024c480a8/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msmqmq "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqmq"
)

// s_DSSetPropsGuidRequest carries the [in] parameters of S_DSSetPropsGuid.
type s_DSSetPropsGuidRequest struct {
	DwObjectType ndr.DWORD
	PGuid        dtyp.GUID
	Cp           ndr.DWORD
	AProp        []ndr.DWORD          `ndr:"ref,size_is=Cp"`
	ApVar        []msmqmq.PROPVARIANT `ndr:"ref,size_is=Cp"`
}

func (*s_DSSetPropsGuidRequest) Opnum() uint16 { return dscomm.OpnumS_DSSetPropsGuid }

// s_DSSetPropsGuidResponse carries the [out] parameters and return value of S_DSSetPropsGuid.
type s_DSSetPropsGuidResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// S_DSSetPropsGuid calls S_DSSetPropsGuid (opnum 12) ([MS-MQDS] — verify the parameter
// modeling and status handling).
func S_DSSetPropsGuid(rpc ndr.Invoker, dwObjectType ndr.DWORD, pGuid dtyp.GUID, cp ndr.DWORD, aProp []ndr.DWORD, apVar []msmqmq.PROPVARIANT) (err error) {
	req := &s_DSSetPropsGuidRequest{
		DwObjectType: dwObjectType,
		PGuid:        pGuid,
		Cp:           cp,
		AProp:        aProp,
		ApVar:        apVar,
	}
	var resp s_DSSetPropsGuidResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("S_DSSetPropsGuid: %w", err)
		return
	}
	if uint32(resp.Status) != dscomm.StatusSuccess {
		err = fmt.Errorf("S_DSSetPropsGuid failed: %s", dscomm.StatusString(uint32(resp.Status)))
	}
	return
}
