package functions

// IDL source: [MS-FRS2] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-frs2/39bd498b-2a94-41b7-914e-10921d543912
// A fetched copy is kept at ms-frs2.idl in the interface directory.

import (
	"fmt"

	FrsTransport "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/897e2e5f-93f3-4376-9c9c-fd2277495c27/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msfrs2 "github.com/TheManticoreProject/Manticore/windows/protocols/ms-frs2"
)

// asyncPollRequest carries the [in] parameters of AsyncPoll.
type asyncPollRequest struct {
	ConnectionId msfrs2.FRS_CONNECTION_ID
}

func (*asyncPollRequest) Opnum() uint16 { return FrsTransport.OpnumAsyncPoll }

// asyncPollResponse carries the [out] parameters and return value of AsyncPoll.
type asyncPollResponse struct {
	Response msfrs2.FRS_ASYNC_RESPONSE_CONTEXT
	Status   ndr.DWORD `ndr:"retval"`
}

// AsyncPoll calls AsyncPoll (opnum 5) ([MS-FRS2] — verify the parameter
// modeling and status handling).
func AsyncPoll(rpc ndr.Invoker, connectionId msfrs2.FRS_CONNECTION_ID) (Response msfrs2.FRS_ASYNC_RESPONSE_CONTEXT, err error) {
	req := &asyncPollRequest{
		ConnectionId: connectionId,
	}
	var resp asyncPollResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("AsyncPoll: %w", err)
		return
	}
	Response = resp.Response
	if uint32(resp.Status) != FrsTransport.StatusSuccess {
		err = fmt.Errorf("AsyncPoll failed: %s", FrsTransport.StatusString(uint32(resp.Status)))
	}
	return
}
