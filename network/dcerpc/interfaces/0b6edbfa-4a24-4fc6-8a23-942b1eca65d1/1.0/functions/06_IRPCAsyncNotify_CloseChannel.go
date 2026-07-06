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

// iRPCAsyncNotify_CloseChannelRequest carries the [in]/[in,out] parameters of
// IRPCAsyncNotify_CloseChannel. pChannel is an [in,out] channel context handle;
// pInNotificationType is a required (top-level [ref]) pointer carried inline; pReason is
// [in, size_is(InSize), unique] byte* — a [unique] pointer to a conformant byte array.
type iRPCAsyncNotify_CloseChannelRequest struct {
	PChannel            mspan.PNOTIFYOBJECT
	PInNotificationType mspan.PrintAsyncNotificationType
	InSize              ndr.DWORD
	PReason             []byte `ndr:"unique,size_is=InSize"`
}

func (*iRPCAsyncNotify_CloseChannelRequest) Opnum() uint16 {
	return IRPCAsyncNotify.OpnumIRPCAsyncNotify_CloseChannel
}

// iRPCAsyncNotify_CloseChannelResponse carries the [out] parameters and return value of IRPCAsyncNotify_CloseChannel.
type iRPCAsyncNotify_CloseChannelResponse struct {
	PChannel mspan.PNOTIFYOBJECT
	Status   ndr.DWORD `ndr:"retval"`
}

// IRPCAsyncNotify_CloseChannel calls IRPCAsyncNotify_CloseChannel (opnum 6)
// ([MS-PAN] 3.1.4.6). It closes a bidirectional notification channel, optionally
// supplying a reason blob to the server.
func IRPCAsyncNotify_CloseChannel(rpc ndr.Invoker, pChannel mspan.PNOTIFYOBJECT, pInNotificationType mspan.PrintAsyncNotificationType, inSize ndr.DWORD, pReason []byte) (PChannel mspan.PNOTIFYOBJECT, err error) {
	req := &iRPCAsyncNotify_CloseChannelRequest{
		PChannel:            pChannel,
		PInNotificationType: pInNotificationType,
		InSize:              inSize,
		PReason:             pReason,
	}
	var resp iRPCAsyncNotify_CloseChannelResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("IRPCAsyncNotify_CloseChannel: %w", err)
		return
	}
	PChannel = resp.PChannel
	if uint32(resp.Status) != IRPCAsyncNotify.StatusSuccess {
		err = fmt.Errorf("IRPCAsyncNotify_CloseChannel failed: %s", IRPCAsyncNotify.StatusString(uint32(resp.Status)))
	}
	return
}
