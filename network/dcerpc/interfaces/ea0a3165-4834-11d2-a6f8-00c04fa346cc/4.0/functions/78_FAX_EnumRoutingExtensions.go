package functions

import (
	"fmt"

	fax "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/ea0a3165-4834-11d2-a6f8-00c04fa346cc/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// fAX_EnumRoutingExtensionsRequest carries the [in] parameters of FAX_EnumRoutingExtensions.
type fAX_EnumRoutingExtensionsRequest struct {
}

func (*fAX_EnumRoutingExtensionsRequest) Opnum() uint16 { return fax.OpnumFAX_EnumRoutingExtensions }

// fAX_EnumRoutingExtensionsResponse carries the [out] parameters and return value of FAX_EnumRoutingExtensions.
type fAX_EnumRoutingExtensionsResponse struct {
	Buffer      []byte `ndr:"unique,conformant"`
	BufferSize  ndr.DWORD
	LpdwNumExts ndr.DWORD
	Status      ndr.DWORD `ndr:"retval"`
}

// FAX_EnumRoutingExtensions calls FAX_EnumRoutingExtensions (opnum 78) ([MS-FAX] — verify the parameter
// modeling and status handling).
func FAX_EnumRoutingExtensions(rpc ndr.Invoker) (Buffer []byte, BufferSize ndr.DWORD, LpdwNumExts ndr.DWORD, err error) {
	req := &fAX_EnumRoutingExtensionsRequest{}
	var resp fAX_EnumRoutingExtensionsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FAX_EnumRoutingExtensions: %w", err)
		return
	}
	Buffer = resp.Buffer
	BufferSize = resp.BufferSize
	LpdwNumExts = resp.LpdwNumExts
	if uint32(resp.Status) != fax.StatusSuccess {
		err = fmt.Errorf("FAX_EnumRoutingExtensions failed: %s", fax.StatusString(uint32(resp.Status)))
	}
	return
}
