package functions

import (
	"fmt"

	IRPCAsyncNotify "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/0b6edbfa-4a24-4fc6-8a23-942b1eca65d1/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mspan "github.com/TheManticoreProject/Manticore/windows/protocols/ms-pan"
)

// iRPCAsyncNotify_RegisterClientRequest carries the [in] parameters of IRPCAsyncNotify_RegisterClient.
//
// pRegistrationObj is a by-value context handle. pName is an optional ([unique])
// UTF-16LE print-queue UNC name; pInNotificationType is a required (top-level [ref])
// pointer to a notification-type GUID, so it is carried inline. NotifyFilter and
// conversationStyle are 32-bit [v1_enum] values.
type iRPCAsyncNotify_RegisterClientRequest struct {
	PRegistrationObj    mspan.PRPCREMOTEOBJECT
	PName               *ndr.WSTR `ndr:"unique"`
	PInNotificationType mspan.PrintAsyncNotificationType
	NotifyFilter        mspan.PrintAsyncNotifyUserFilter
	ConversationStyle   mspan.PrintAsyncNotifyConversationStyle
}

func (*iRPCAsyncNotify_RegisterClientRequest) Opnum() uint16 {
	return IRPCAsyncNotify.OpnumIRPCAsyncNotify_RegisterClient
}

// iRPCAsyncNotify_RegisterClientResponse carries the [out] parameter and HRESULT of
// IRPCAsyncNotify_RegisterClient. ppRmtServerReferral is a [string] wchar_t** ([out]);
// servers SHOULD return NULL and clients MUST ignore it, so it is a [unique] string
// pointer that is normally nil.
type iRPCAsyncNotify_RegisterClientResponse struct {
	PpRmtServerReferral *ndr.WSTR `ndr:"unique"`
	Status              ndr.DWORD `ndr:"retval"`
}

// IRPCAsyncNotify_RegisterClient calls IRPCAsyncNotify_RegisterClient (opnum 0)
// ([MS-PAN] 3.1.4.1). It registers the client for the given notification type, user
// filter, and conversation style against a remote object obtained from
// IRPCRemoteObject_Create.
func IRPCAsyncNotify_RegisterClient(rpc ndr.Invoker, pRegistrationObj mspan.PRPCREMOTEOBJECT, pName *ndr.WSTR, pInNotificationType mspan.PrintAsyncNotificationType, notifyFilter mspan.PrintAsyncNotifyUserFilter, conversationStyle mspan.PrintAsyncNotifyConversationStyle) (PpRmtServerReferral *ndr.WSTR, err error) {
	req := &iRPCAsyncNotify_RegisterClientRequest{
		PRegistrationObj:    pRegistrationObj,
		PName:               pName,
		PInNotificationType: pInNotificationType,
		NotifyFilter:        notifyFilter,
		ConversationStyle:   conversationStyle,
	}
	var resp iRPCAsyncNotify_RegisterClientResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("IRPCAsyncNotify_RegisterClient: %w", err)
		return
	}
	PpRmtServerReferral = resp.PpRmtServerReferral
	if uint32(resp.Status) != IRPCAsyncNotify.StatusSuccess {
		err = fmt.Errorf("IRPCAsyncNotify_RegisterClient failed: %s", IRPCAsyncNotify.StatusString(uint32(resp.Status)))
	}
	return
}
