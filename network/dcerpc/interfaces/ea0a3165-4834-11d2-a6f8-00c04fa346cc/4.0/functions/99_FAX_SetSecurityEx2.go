package functions

import (
	"fmt"

	fax "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/ea0a3165-4834-11d2-a6f8-00c04fa346cc/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// fAX_SetSecurityEx2Request carries the [in] parameters of FAX_SetSecurityEx2.
type fAX_SetSecurityEx2Request struct {
	SecurityInformation ndr.DWORD
	PSecurityDescriptor []uint8 `ndr:"ref,size_is=DwBufferSize"`
	DwBufferSize        ndr.DWORD
}

func (*fAX_SetSecurityEx2Request) Opnum() uint16 { return fax.OpnumFAX_SetSecurityEx2 }

// fAX_SetSecurityEx2Response carries the [out] parameters and return value of FAX_SetSecurityEx2.
type fAX_SetSecurityEx2Response struct {
	Status ndr.DWORD `ndr:"retval"`
}

// FAX_SetSecurityEx2 calls FAX_SetSecurityEx2 (opnum 99) ([MS-FAX] — verify the parameter
// modeling and status handling).
func FAX_SetSecurityEx2(rpc ndr.Invoker, securityInformation ndr.DWORD, pSecurityDescriptor []uint8, dwBufferSize ndr.DWORD) (err error) {
	req := &fAX_SetSecurityEx2Request{
		SecurityInformation: securityInformation,
		PSecurityDescriptor: pSecurityDescriptor,
		DwBufferSize:        dwBufferSize,
	}
	var resp fAX_SetSecurityEx2Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FAX_SetSecurityEx2: %w", err)
		return
	}
	if uint32(resp.Status) != fax.StatusSuccess {
		err = fmt.Errorf("FAX_SetSecurityEx2 failed: %s", fax.StatusString(uint32(resp.Status)))
	}
	return
}
