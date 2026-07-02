package functions

import (
	"fmt"

	fax "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/ea0a3165-4834-11d2-a6f8-00c04fa346cc/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msfax "github.com/TheManticoreProject/Manticore/windows/protocols/ms-fax"
)

// fAX_StartServerNotificationRequest carries the [in] parameters of FAX_StartServerNotification.
type fAX_StartServerNotificationRequest struct {
	LpcwstrMachineName   ndr.WSTR
	LpcwstrEndPoint      ndr.WSTR
	Context              uint64
	LpcwstrProtseqString ndr.WSTR
	BEventEx             ndr.BOOL
	DwEventTypes         ndr.DWORD
}

func (*fAX_StartServerNotificationRequest) Opnum() uint16 {
	return fax.OpnumFAX_StartServerNotification
}

// fAX_StartServerNotificationResponse carries the [out] parameters and return value of FAX_StartServerNotification.
type fAX_StartServerNotificationResponse struct {
	LpHandle msfax.PRPC_FAX_EVENT_HANDLE
	Status   ndr.DWORD `ndr:"retval"`
}

// FAX_StartServerNotification calls FAX_StartServerNotification (opnum 73) ([MS-FAX] — verify the parameter
// modeling and status handling).
func FAX_StartServerNotification(rpc ndr.Invoker, lpcwstrMachineName ndr.WSTR, lpcwstrEndPoint ndr.WSTR, context uint64, lpcwstrProtseqString ndr.WSTR, bEventEx ndr.BOOL, dwEventTypes ndr.DWORD) (LpHandle msfax.PRPC_FAX_EVENT_HANDLE, err error) {
	req := &fAX_StartServerNotificationRequest{
		LpcwstrMachineName:   lpcwstrMachineName,
		LpcwstrEndPoint:      lpcwstrEndPoint,
		Context:              context,
		LpcwstrProtseqString: lpcwstrProtseqString,
		BEventEx:             bEventEx,
		DwEventTypes:         dwEventTypes,
	}
	var resp fAX_StartServerNotificationResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FAX_StartServerNotification: %w", err)
		return
	}
	LpHandle = resp.LpHandle
	if uint32(resp.Status) != fax.StatusSuccess {
		err = fmt.Errorf("FAX_StartServerNotification failed: %s", fax.StatusString(uint32(resp.Status)))
	}
	return
}
