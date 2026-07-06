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

// fAX_GetSecurityEx2Request carries the [in] parameters of FAX_GetSecurityEx2.
type fAX_GetSecurityEx2Request struct {
	SecurityInformation ndr.DWORD
}

func (*fAX_GetSecurityEx2Request) Opnum() uint16 { return fax.OpnumFAX_GetSecurityEx2 }

// fAX_GetSecurityEx2Response carries the [out] parameters and return value of FAX_GetSecurityEx2.
type fAX_GetSecurityEx2Response struct {
	PSecurityDescriptor []byte `ndr:"unique,conformant"`
	LpdwBufferSize      ndr.DWORD
	Status              ndr.DWORD `ndr:"retval"`
}

// FAX_GetSecurityEx2 calls FAX_GetSecurityEx2 (opnum 98) ([MS-FAX] — verify the parameter
// modeling and status handling).
func FAX_GetSecurityEx2(rpc ndr.Invoker, securityInformation ndr.DWORD) (PSecurityDescriptor []byte, LpdwBufferSize ndr.DWORD, err error) {
	req := &fAX_GetSecurityEx2Request{
		SecurityInformation: securityInformation,
	}
	var resp fAX_GetSecurityEx2Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FAX_GetSecurityEx2: %w", err)
		return
	}
	PSecurityDescriptor = resp.PSecurityDescriptor
	LpdwBufferSize = resp.LpdwBufferSize
	if uint32(resp.Status) != fax.StatusSuccess {
		err = fmt.Errorf("FAX_GetSecurityEx2 failed: %s", fax.StatusString(uint32(resp.Status)))
	}
	return
}
