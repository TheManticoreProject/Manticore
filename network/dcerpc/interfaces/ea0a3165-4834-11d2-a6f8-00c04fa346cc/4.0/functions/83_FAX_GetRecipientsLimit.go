package functions

import (
	"fmt"

	fax "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/ea0a3165-4834-11d2-a6f8-00c04fa346cc/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// fAX_GetRecipientsLimitRequest carries the [in] parameters of FAX_GetRecipientsLimit.
type fAX_GetRecipientsLimitRequest struct {
}

func (*fAX_GetRecipientsLimitRequest) Opnum() uint16 { return fax.OpnumFAX_GetRecipientsLimit }

// fAX_GetRecipientsLimitResponse carries the [out] parameters and return value of FAX_GetRecipientsLimit.
type fAX_GetRecipientsLimitResponse struct {
	LpdwRecipientsLimit ndr.DWORD
	Status              ndr.DWORD `ndr:"retval"`
}

// FAX_GetRecipientsLimit calls FAX_GetRecipientsLimit (opnum 83) ([MS-FAX] — verify the parameter
// modeling and status handling).
func FAX_GetRecipientsLimit(rpc ndr.Invoker) (LpdwRecipientsLimit ndr.DWORD, err error) {
	req := &fAX_GetRecipientsLimitRequest{}
	var resp fAX_GetRecipientsLimitResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FAX_GetRecipientsLimit: %w", err)
		return
	}
	LpdwRecipientsLimit = resp.LpdwRecipientsLimit
	if uint32(resp.Status) != fax.StatusSuccess {
		err = fmt.Errorf("FAX_GetRecipientsLimit failed: %s", fax.StatusString(uint32(resp.Status)))
	}
	return
}
