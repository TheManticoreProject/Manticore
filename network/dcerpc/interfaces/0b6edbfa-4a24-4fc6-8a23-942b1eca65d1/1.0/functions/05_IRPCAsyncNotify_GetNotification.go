package functions

import (
	"fmt"

	IRPCAsyncNotify "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/0b6edbfa-4a24-4fc6-8a23-942b1eca65d1/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mspan "github.com/TheManticoreProject/Manticore/windows/protocols/ms-pan"
)

// iRPCAsyncNotify_GetNotificationRequest carries the [in] parameters of IRPCAsyncNotify_GetNotification.
type iRPCAsyncNotify_GetNotificationRequest struct {
	PRemoteObj mspan.PRPCREMOTEOBJECT
}

func (*iRPCAsyncNotify_GetNotificationRequest) Opnum() uint16 {
	return IRPCAsyncNotify.OpnumIRPCAsyncNotify_GetNotification
}

// iRPCAsyncNotify_GetNotificationResponse carries the [out] parameters and HRESULT of
// IRPCAsyncNotify_GetNotification. ppOutNotificationData is [out, size_is( , *pOutSize)]
// byte**: the top-level [out] indirection, then a [unique] pointer to a conformant array
// of POutSize octets.
type iRPCAsyncNotify_GetNotificationResponse struct {
	PpOutNotificationType *mspan.PrintAsyncNotificationType `ndr:"unique"`
	POutSize              ndr.DWORD
	PpOutNotificationData []byte    `ndr:"unique,size_is=POutSize"`
	Status                ndr.DWORD `ndr:"retval"`
}

// IRPCAsyncNotify_GetNotification calls IRPCAsyncNotify_GetNotification (opnum 5)
// ([MS-PAN] 3.1.4.5). For a unidirectional registration it blocks until the next
// notification is available and returns its type and data.
func IRPCAsyncNotify_GetNotification(rpc ndr.Invoker, pRemoteObj mspan.PRPCREMOTEOBJECT) (PpOutNotificationType *mspan.PrintAsyncNotificationType, POutSize ndr.DWORD, PpOutNotificationData []byte, err error) {
	req := &iRPCAsyncNotify_GetNotificationRequest{
		PRemoteObj: pRemoteObj,
	}
	var resp iRPCAsyncNotify_GetNotificationResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("IRPCAsyncNotify_GetNotification: %w", err)
		return
	}
	PpOutNotificationType = resp.PpOutNotificationType
	POutSize = resp.POutSize
	PpOutNotificationData = resp.PpOutNotificationData
	if uint32(resp.Status) != IRPCAsyncNotify.StatusSuccess {
		err = fmt.Errorf("IRPCAsyncNotify_GetNotification failed: %s", IRPCAsyncNotify.StatusString(uint32(resp.Status)))
	}
	return
}
