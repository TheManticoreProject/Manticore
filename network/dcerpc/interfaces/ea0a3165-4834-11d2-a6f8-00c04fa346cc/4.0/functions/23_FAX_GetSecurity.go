package functions

import (
	"fmt"

	fax "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/ea0a3165-4834-11d2-a6f8-00c04fa346cc/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// fAX_GetSecurityRequest carries the [in] parameters of FAX_GetSecurity.
type fAX_GetSecurityRequest struct {
}

func (*fAX_GetSecurityRequest) Opnum() uint16 { return fax.OpnumFAX_GetSecurity }

// fAX_GetSecurityResponse carries the [out] parameters and return value of FAX_GetSecurity.
type fAX_GetSecurityResponse struct {
	PSecurityDescriptor []byte `ndr:"unique,conformant"`
	LpdwBufferSize      ndr.DWORD
	Status              ndr.DWORD `ndr:"retval"`
}

// FAX_GetSecurity calls FAX_GetSecurity (opnum 23) ([MS-FAX] — verify the parameter
// modeling and status handling).
func FAX_GetSecurity(rpc ndr.Invoker) (PSecurityDescriptor []byte, LpdwBufferSize ndr.DWORD, err error) {
	req := &fAX_GetSecurityRequest{}
	var resp fAX_GetSecurityResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FAX_GetSecurity: %w", err)
		return
	}
	PSecurityDescriptor = resp.PSecurityDescriptor
	LpdwBufferSize = resp.LpdwBufferSize
	if uint32(resp.Status) != fax.StatusSuccess {
		err = fmt.Errorf("FAX_GetSecurity failed: %s", fax.StatusString(uint32(resp.Status)))
	}
	return
}
