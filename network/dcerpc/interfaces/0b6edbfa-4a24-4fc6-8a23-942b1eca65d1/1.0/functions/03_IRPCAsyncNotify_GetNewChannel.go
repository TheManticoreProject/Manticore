package functions

import (
	"fmt"

	IRPCAsyncNotify "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/0b6edbfa-4a24-4fc6-8a23-942b1eca65d1/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mspan "github.com/TheManticoreProject/Manticore/windows/protocols/ms-pan"
)

// iRPCAsyncNotify_GetNewChannelRequest carries the [in] parameters of IRPCAsyncNotify_GetNewChannel.
type iRPCAsyncNotify_GetNewChannelRequest struct {
	PRemoteObj mspan.PRPCREMOTEOBJECT
}

func (*iRPCAsyncNotify_GetNewChannelRequest) Opnum() uint16 {
	return IRPCAsyncNotify.OpnumIRPCAsyncNotify_GetNewChannel
}

// iRPCAsyncNotify_GetNewChannelResponse carries the [out] parameters and HRESULT of
// IRPCAsyncNotify_GetNewChannel. ppChannelCtxt is [out, size_is( , *pNoOfChannels)]
// PNOTIFYOBJECT**: the top-level [out] indirection, then a [unique] pointer to a
// conformant array of PNoOfChannels context handles (each carried inline as 20 octets).
type iRPCAsyncNotify_GetNewChannelResponse struct {
	PNoOfChannels ndr.DWORD
	PpChannelCtxt []mspan.PNOTIFYOBJECT `ndr:"unique,size_is=PNoOfChannels"`
	Status        ndr.DWORD             `ndr:"retval"`
}

// IRPCAsyncNotify_GetNewChannel calls IRPCAsyncNotify_GetNewChannel (opnum 3)
// ([MS-PAN] 3.1.4.3). For a bidirectional registration it returns one or more new
// notification-channel context handles.
func IRPCAsyncNotify_GetNewChannel(rpc ndr.Invoker, pRemoteObj mspan.PRPCREMOTEOBJECT) (PNoOfChannels ndr.DWORD, PpChannelCtxt []mspan.PNOTIFYOBJECT, err error) {
	req := &iRPCAsyncNotify_GetNewChannelRequest{
		PRemoteObj: pRemoteObj,
	}
	var resp iRPCAsyncNotify_GetNewChannelResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("IRPCAsyncNotify_GetNewChannel: %w", err)
		return
	}
	PNoOfChannels = resp.PNoOfChannels
	PpChannelCtxt = resp.PpChannelCtxt
	if uint32(resp.Status) != IRPCAsyncNotify.StatusSuccess {
		err = fmt.Errorf("IRPCAsyncNotify_GetNewChannel failed: %s", IRPCAsyncNotify.StatusString(uint32(resp.Status)))
	}
	return
}
