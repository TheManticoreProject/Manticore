package functions

// IDL source: [MS-FAX] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-fax/0907310f-0693-47e7-a6cb-3e599c89a1dd
// A fetched copy is kept at ms-fax.idl in the interface directory.

import (
	"fmt"

	fax "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/ea0a3165-4834-11d2-a6f8-00c04fa346cc/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msfax "github.com/TheManticoreProject/Manticore/windows/protocols/ms-fax"
)

// fAX_EndServerNotificationRequest carries the [in] parameters of FAX_EndServerNotification.
type fAX_EndServerNotificationRequest struct {
	LpHandle msfax.PRPC_FAX_EVENT_EX_HANDLE
}

func (*fAX_EndServerNotificationRequest) Opnum() uint16 { return fax.OpnumFAX_EndServerNotification }

// fAX_EndServerNotificationResponse carries the [out] parameters and return value of FAX_EndServerNotification.
type fAX_EndServerNotificationResponse struct {
	LpHandle msfax.PRPC_FAX_EVENT_EX_HANDLE
	Status   ndr.DWORD `ndr:"retval"`
}

// FAX_EndServerNotification calls FAX_EndServerNotification (opnum 75) ([MS-FAX] — verify the parameter
// modeling and status handling).
func FAX_EndServerNotification(rpc ndr.Invoker, lpHandle msfax.PRPC_FAX_EVENT_EX_HANDLE) (LpHandle msfax.PRPC_FAX_EVENT_EX_HANDLE, err error) {
	req := &fAX_EndServerNotificationRequest{
		LpHandle: lpHandle,
	}
	var resp fAX_EndServerNotificationResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FAX_EndServerNotification: %w", err)
		return
	}
	LpHandle = resp.LpHandle
	if uint32(resp.Status) != fax.StatusSuccess {
		err = fmt.Errorf("FAX_EndServerNotification failed: %s", fax.StatusString(uint32(resp.Status)))
	}
	return
}
