package functions

import (
	"fmt"

	fax "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/ea0a3165-4834-11d2-a6f8-00c04fa346cc/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// fAX_UnregisterServiceProviderExRequest carries the [in] parameters of FAX_UnregisterServiceProviderEx.
type fAX_UnregisterServiceProviderExRequest struct {
	LpcwstrGUID ndr.WSTR
}

func (*fAX_UnregisterServiceProviderExRequest) Opnum() uint16 {
	return fax.OpnumFAX_UnregisterServiceProviderEx
}

// fAX_UnregisterServiceProviderExResponse carries the [out] parameters and return value of FAX_UnregisterServiceProviderEx.
type fAX_UnregisterServiceProviderExResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// FAX_UnregisterServiceProviderEx calls FAX_UnregisterServiceProviderEx (opnum 61) ([MS-FAX] — verify the parameter
// modeling and status handling).
func FAX_UnregisterServiceProviderEx(rpc ndr.Invoker, lpcwstrGUID ndr.WSTR) (err error) {
	req := &fAX_UnregisterServiceProviderExRequest{
		LpcwstrGUID: lpcwstrGUID,
	}
	var resp fAX_UnregisterServiceProviderExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FAX_UnregisterServiceProviderEx: %w", err)
		return
	}
	if uint32(resp.Status) != fax.StatusSuccess {
		err = fmt.Errorf("FAX_UnregisterServiceProviderEx failed: %s", fax.StatusString(uint32(resp.Status)))
	}
	return
}
