package functions

// IDL source: [MS-PAN] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-pan/3161e1b8-098f-4f42-8a58-7e342114b643
// A fetched copy is kept at ms-pan.idl in the interface directory.

import (
	"fmt"

	IRPCAsyncNotify "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/0b6edbfa-4a24-4fc6-8a23-942b1eca65d1/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mspan "github.com/TheManticoreProject/Manticore/windows/protocols/ms-pan"
)

// iRPCAsyncNotify_GetNotificationSendResponseRequest carries the [in]/[in,out] parameters
// of IRPCAsyncNotify_GetNotificationSendResponse. pChannel is an [in,out] channel context
// handle; pInNotificationData is [in, size_is(InSize), unique] byte* — a [unique] pointer
// to a conformant byte array of InSize octets.
type iRPCAsyncNotify_GetNotificationSendResponseRequest struct {
	PChannel            mspan.PNOTIFYOBJECT
	PInNotificationType *mspan.PrintAsyncNotificationType `ndr:"unique"`
	InSize              ndr.DWORD
	PInNotificationData []byte `ndr:"unique,size_is=InSize"`
}

func (*iRPCAsyncNotify_GetNotificationSendResponseRequest) Opnum() uint16 {
	return IRPCAsyncNotify.OpnumIRPCAsyncNotify_GetNotificationSendResponse
}

// iRPCAsyncNotify_GetNotificationSendResponseResponse carries the [in,out]/[out] parameters
// and HRESULT. ppOutNotificationData is [out, size_is( , *pOutSize)] byte**: the top-level
// [out] indirection, then a [unique] pointer to a conformant array of POutSize octets.
type iRPCAsyncNotify_GetNotificationSendResponseResponse struct {
	PChannel              mspan.PNOTIFYOBJECT
	PpOutNotificationType *mspan.PrintAsyncNotificationType `ndr:"unique"`
	POutSize              ndr.DWORD
	PpOutNotificationData []byte    `ndr:"unique,size_is=POutSize"`
	Status                ndr.DWORD `ndr:"retval"`
}

// IRPCAsyncNotify_GetNotificationSendResponse calls IRPCAsyncNotify_GetNotificationSendResponse
// (opnum 4) ([MS-PAN] 3.1.4.4). On a bidirectional channel it sends the client's response
// data and blocks until the next notification (or channel closure) is returned.
func IRPCAsyncNotify_GetNotificationSendResponse(rpc ndr.Invoker, pChannel mspan.PNOTIFYOBJECT, pInNotificationType *mspan.PrintAsyncNotificationType, inSize ndr.DWORD, pInNotificationData []byte) (PChannel mspan.PNOTIFYOBJECT, PpOutNotificationType *mspan.PrintAsyncNotificationType, POutSize ndr.DWORD, PpOutNotificationData []byte, err error) {
	req := &iRPCAsyncNotify_GetNotificationSendResponseRequest{
		PChannel:            pChannel,
		PInNotificationType: pInNotificationType,
		InSize:              inSize,
		PInNotificationData: pInNotificationData,
	}
	var resp iRPCAsyncNotify_GetNotificationSendResponseResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("IRPCAsyncNotify_GetNotificationSendResponse: %w", err)
		return
	}
	PChannel = resp.PChannel
	PpOutNotificationType = resp.PpOutNotificationType
	POutSize = resp.POutSize
	PpOutNotificationData = resp.PpOutNotificationData
	if uint32(resp.Status) != IRPCAsyncNotify.StatusSuccess {
		err = fmt.Errorf("IRPCAsyncNotify_GetNotificationSendResponse failed: %s", IRPCAsyncNotify.StatusString(uint32(resp.Status)))
	}
	return
}
