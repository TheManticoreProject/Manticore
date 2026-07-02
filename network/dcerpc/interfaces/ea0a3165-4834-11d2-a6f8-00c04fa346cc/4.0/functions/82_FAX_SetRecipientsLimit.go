package functions

import (
	"fmt"

	fax "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/ea0a3165-4834-11d2-a6f8-00c04fa346cc/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// fAX_SetRecipientsLimitRequest carries the [in] parameters of FAX_SetRecipientsLimit.
type fAX_SetRecipientsLimitRequest struct {
	DwRecipientsLimit ndr.DWORD
}

func (*fAX_SetRecipientsLimitRequest) Opnum() uint16 { return fax.OpnumFAX_SetRecipientsLimit }

// fAX_SetRecipientsLimitResponse carries the [out] parameters and return value of FAX_SetRecipientsLimit.
type fAX_SetRecipientsLimitResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// FAX_SetRecipientsLimit calls FAX_SetRecipientsLimit (opnum 82) ([MS-FAX] — verify the parameter
// modeling and status handling).
func FAX_SetRecipientsLimit(rpc ndr.Invoker, dwRecipientsLimit ndr.DWORD) (err error) {
	req := &fAX_SetRecipientsLimitRequest{
		DwRecipientsLimit: dwRecipientsLimit,
	}
	var resp fAX_SetRecipientsLimitResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FAX_SetRecipientsLimit: %w", err)
		return
	}
	if uint32(resp.Status) != fax.StatusSuccess {
		err = fmt.Errorf("FAX_SetRecipientsLimit failed: %s", fax.StatusString(uint32(resp.Status)))
	}
	return
}
