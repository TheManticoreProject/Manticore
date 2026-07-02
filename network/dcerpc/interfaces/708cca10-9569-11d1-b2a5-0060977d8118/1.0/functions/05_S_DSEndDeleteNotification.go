package functions

import (
	"fmt"

	dscomm2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/708cca10-9569-11d1-b2a5-0060977d8118/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msmqds "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqds"
)

// s_DSEndDeleteNotificationRequest carries the [in] parameters of S_DSEndDeleteNotification.
type s_DSEndDeleteNotificationRequest struct {
	PHandle msmqds.PPCONTEXT_HANDLE_DELETE_TYPE
}

func (*s_DSEndDeleteNotificationRequest) Opnum() uint16 {
	return dscomm2.OpnumS_DSEndDeleteNotification
}

// s_DSEndDeleteNotificationResponse carries the [in, out] context handle of
// S_DSEndDeleteNotification. The method returns void ([MS-MQDS] 3.1.4.6): there is no
// HRESULT on the wire, so the response is just the (nulled) handle.
type s_DSEndDeleteNotificationResponse struct {
	PHandle msmqds.PPCONTEXT_HANDLE_DELETE_TYPE
}

// S_DSEndDeleteNotification calls S_DSEndDeleteNotification (opnum 5) and returns the
// server-updated context handle ([MS-MQDS] 3.1.4.6). The method has no return value.
func S_DSEndDeleteNotification(rpc ndr.Invoker, pHandle msmqds.PPCONTEXT_HANDLE_DELETE_TYPE) (PHandle msmqds.PPCONTEXT_HANDLE_DELETE_TYPE, err error) {
	req := &s_DSEndDeleteNotificationRequest{
		PHandle: pHandle,
	}
	var resp s_DSEndDeleteNotificationResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("S_DSEndDeleteNotification: %w", err)
		return
	}
	PHandle = resp.PHandle
	return
}
