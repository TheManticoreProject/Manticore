package functions

// IDL source: [MS-FAX] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-fax/0907310f-0693-47e7-a6cb-3e599c89a1dd
// A fetched copy is kept at ms-fax.idl in the interface directory.

import (
	"fmt"

	fax "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/ea0a3165-4834-11d2-a6f8-00c04fa346cc/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// fAX_SetSecurityRequest carries the [in] parameters of FAX_SetSecurity.
type fAX_SetSecurityRequest struct {
	SecurityInformation ndr.DWORD
	PSecurityDescriptor []uint8 `ndr:"ref,size_is=DwBufferSize"`
	DwBufferSize        ndr.DWORD
}

func (*fAX_SetSecurityRequest) Opnum() uint16 { return fax.OpnumFAX_SetSecurity }

// fAX_SetSecurityResponse carries the [out] parameters and return value of FAX_SetSecurity.
type fAX_SetSecurityResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// FAX_SetSecurity calls FAX_SetSecurity (opnum 24) ([MS-FAX] — verify the parameter
// modeling and status handling).
func FAX_SetSecurity(rpc ndr.Invoker, securityInformation ndr.DWORD, pSecurityDescriptor []uint8, dwBufferSize ndr.DWORD) (err error) {
	req := &fAX_SetSecurityRequest{
		SecurityInformation: securityInformation,
		PSecurityDescriptor: pSecurityDescriptor,
		DwBufferSize:        dwBufferSize,
	}
	var resp fAX_SetSecurityResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FAX_SetSecurity: %w", err)
		return
	}
	if uint32(resp.Status) != fax.StatusSuccess {
		err = fmt.Errorf("FAX_SetSecurity failed: %s", fax.StatusString(uint32(resp.Status)))
	}
	return
}
