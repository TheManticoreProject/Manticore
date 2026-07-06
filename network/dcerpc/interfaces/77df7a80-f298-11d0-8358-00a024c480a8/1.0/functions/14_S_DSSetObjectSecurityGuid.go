package functions

import (
	"fmt"

	dscomm "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/77df7a80-f298-11d0-8358-00a024c480a8/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// s_DSSetObjectSecurityGuidRequest carries the [in] parameters of S_DSSetObjectSecurityGuid.
type s_DSSetObjectSecurityGuidRequest struct {
	DwObjectType        ndr.DWORD
	PGuid               msdtyp.GUID
	SecurityInformation ndr.DWORD
	PSecurityDescriptor []uint8 `ndr:"ref,size_is=NLength"`
	NLength             ndr.DWORD
}

func (*s_DSSetObjectSecurityGuidRequest) Opnum() uint16 { return dscomm.OpnumS_DSSetObjectSecurityGuid }

// s_DSSetObjectSecurityGuidResponse carries the [out] parameters and return value of S_DSSetObjectSecurityGuid.
type s_DSSetObjectSecurityGuidResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// S_DSSetObjectSecurityGuid calls S_DSSetObjectSecurityGuid (opnum 14) ([MS-MQDS] — verify the parameter
// modeling and status handling).
func S_DSSetObjectSecurityGuid(rpc ndr.Invoker, dwObjectType ndr.DWORD, pGuid msdtyp.GUID, securityInformation ndr.DWORD, pSecurityDescriptor []uint8, nLength ndr.DWORD) (err error) {
	req := &s_DSSetObjectSecurityGuidRequest{
		DwObjectType:        dwObjectType,
		PGuid:               pGuid,
		SecurityInformation: securityInformation,
		PSecurityDescriptor: pSecurityDescriptor,
		NLength:             nLength,
	}
	var resp s_DSSetObjectSecurityGuidResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("S_DSSetObjectSecurityGuid: %w", err)
		return
	}
	if uint32(resp.Status) != dscomm.StatusSuccess {
		err = fmt.Errorf("S_DSSetObjectSecurityGuid failed: %s", dscomm.StatusString(uint32(resp.Status)))
	}
	return
}
