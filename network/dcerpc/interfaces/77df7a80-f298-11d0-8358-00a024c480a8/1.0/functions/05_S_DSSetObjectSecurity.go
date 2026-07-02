package functions

import (
	"fmt"

	dscomm "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/77df7a80-f298-11d0-8358-00a024c480a8/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// s_DSSetObjectSecurityRequest carries the [in] parameters of S_DSSetObjectSecurity.
type s_DSSetObjectSecurityRequest struct {
	DwObjectType        ndr.DWORD
	PwcsPathName        ndr.WSTR
	SecurityInformation ndr.DWORD
	PSecurityDescriptor []uint8 `ndr:"ref,size_is=NLength"`
	NLength             ndr.DWORD
}

func (*s_DSSetObjectSecurityRequest) Opnum() uint16 { return dscomm.OpnumS_DSSetObjectSecurity }

// s_DSSetObjectSecurityResponse carries the [out] parameters and return value of S_DSSetObjectSecurity.
type s_DSSetObjectSecurityResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// S_DSSetObjectSecurity calls S_DSSetObjectSecurity (opnum 5) ([MS-MQDS] — verify the parameter
// modeling and status handling).
func S_DSSetObjectSecurity(rpc ndr.Invoker, dwObjectType ndr.DWORD, pwcsPathName ndr.WSTR, securityInformation ndr.DWORD, pSecurityDescriptor []uint8, nLength ndr.DWORD) (err error) {
	req := &s_DSSetObjectSecurityRequest{
		DwObjectType:        dwObjectType,
		PwcsPathName:        pwcsPathName,
		SecurityInformation: securityInformation,
		PSecurityDescriptor: pSecurityDescriptor,
		NLength:             nLength,
	}
	var resp s_DSSetObjectSecurityResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("S_DSSetObjectSecurity: %w", err)
		return
	}
	if uint32(resp.Status) != dscomm.StatusSuccess {
		err = fmt.Errorf("S_DSSetObjectSecurity failed: %s", dscomm.StatusString(uint32(resp.Status)))
	}
	return
}
