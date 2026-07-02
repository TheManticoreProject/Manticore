package functions

import (
	"fmt"

	fax "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/ea0a3165-4834-11d2-a6f8-00c04fa346cc/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// fAX_UnregisterRoutingExtensionRequest carries the [in] parameters of FAX_UnregisterRoutingExtension.
type fAX_UnregisterRoutingExtensionRequest struct {
	LpcwstrExtensionName ndr.WSTR
}

func (*fAX_UnregisterRoutingExtensionRequest) Opnum() uint16 {
	return fax.OpnumFAX_UnregisterRoutingExtension
}

// fAX_UnregisterRoutingExtensionResponse carries the [out] parameters and return value of FAX_UnregisterRoutingExtension.
type fAX_UnregisterRoutingExtensionResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// FAX_UnregisterRoutingExtension calls FAX_UnregisterRoutingExtension (opnum 62) ([MS-FAX] — verify the parameter
// modeling and status handling).
func FAX_UnregisterRoutingExtension(rpc ndr.Invoker, lpcwstrExtensionName ndr.WSTR) (err error) {
	req := &fAX_UnregisterRoutingExtensionRequest{
		LpcwstrExtensionName: lpcwstrExtensionName,
	}
	var resp fAX_UnregisterRoutingExtensionResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FAX_UnregisterRoutingExtension: %w", err)
		return
	}
	if uint32(resp.Status) != fax.StatusSuccess {
		err = fmt.Errorf("FAX_UnregisterRoutingExtension failed: %s", fax.StatusString(uint32(resp.Status)))
	}
	return
}
