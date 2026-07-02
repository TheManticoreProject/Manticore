package functions

import (
	"fmt"

	fax "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/ea0a3165-4834-11d2-a6f8-00c04fa346cc/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// fAX_GetReceiptsOptionsRequest carries the [in] parameters of FAX_GetReceiptsOptions.
type fAX_GetReceiptsOptionsRequest struct {
}

func (*fAX_GetReceiptsOptionsRequest) Opnum() uint16 { return fax.OpnumFAX_GetReceiptsOptions }

// fAX_GetReceiptsOptionsResponse carries the [out] parameters and return value of FAX_GetReceiptsOptions.
type fAX_GetReceiptsOptionsResponse struct {
	LpdwReceiptsOptions ndr.DWORD
	Status              ndr.DWORD `ndr:"retval"`
}

// FAX_GetReceiptsOptions calls FAX_GetReceiptsOptions (opnum 36) ([MS-FAX] — verify the parameter
// modeling and status handling).
func FAX_GetReceiptsOptions(rpc ndr.Invoker) (LpdwReceiptsOptions ndr.DWORD, err error) {
	req := &fAX_GetReceiptsOptionsRequest{}
	var resp fAX_GetReceiptsOptionsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FAX_GetReceiptsOptions: %w", err)
		return
	}
	LpdwReceiptsOptions = resp.LpdwReceiptsOptions
	if uint32(resp.Status) != fax.StatusSuccess {
		err = fmt.Errorf("FAX_GetReceiptsOptions failed: %s", fax.StatusString(uint32(resp.Status)))
	}
	return
}
