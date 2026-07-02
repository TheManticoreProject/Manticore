package functions

import (
	"fmt"

	fax "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/ea0a3165-4834-11d2-a6f8-00c04fa346cc/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msfax "github.com/TheManticoreProject/Manticore/windows/protocols/ms-fax"
)

// fAX_SetReceiptsConfigurationRequest carries the [in] parameters of FAX_SetReceiptsConfiguration.
type fAX_SetReceiptsConfigurationRequest struct {
	PReceipts msfax.FAX_RECEIPTS_CONFIGW
}

func (*fAX_SetReceiptsConfigurationRequest) Opnum() uint16 {
	return fax.OpnumFAX_SetReceiptsConfiguration
}

// fAX_SetReceiptsConfigurationResponse carries the [out] parameters and return value of FAX_SetReceiptsConfiguration.
type fAX_SetReceiptsConfigurationResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// FAX_SetReceiptsConfiguration calls FAX_SetReceiptsConfiguration (opnum 35) ([MS-FAX] — verify the parameter
// modeling and status handling).
func FAX_SetReceiptsConfiguration(rpc ndr.Invoker, pReceipts msfax.FAX_RECEIPTS_CONFIGW) (err error) {
	req := &fAX_SetReceiptsConfigurationRequest{
		PReceipts: pReceipts,
	}
	var resp fAX_SetReceiptsConfigurationResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FAX_SetReceiptsConfiguration: %w", err)
		return
	}
	if uint32(resp.Status) != fax.StatusSuccess {
		err = fmt.Errorf("FAX_SetReceiptsConfiguration failed: %s", fax.StatusString(uint32(resp.Status)))
	}
	return
}
