package functions

// IDL source: [MS-TSTS] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-tsts/c43addc7-eebc-491b-9b01-2587262675e8
// A fetched copy is kept at ms-tsts.idl in the interface directory.

import (
	"fmt"

	IcaApi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5ca4a760-ebb1-11cf-8611-00a0245420ed/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mststs "github.com/TheManticoreProject/Manticore/windows/protocols/ms-tsts"
)

// rpcWinStationWaitSystemEventRequest carries the [in] parameters of RpcWinStationWaitSystemEvent.
type rpcWinStationWaitSystemEventRequest struct {
	HServer   mststs.SERVER_HANDLE
	EventMask ndr.DWORD
}

func (*rpcWinStationWaitSystemEventRequest) Opnum() uint16 {
	return IcaApi.OpnumRpcWinStationWaitSystemEvent
}

// rpcWinStationWaitSystemEventResponse carries the [out] parameters and return value of RpcWinStationWaitSystemEvent.
type rpcWinStationWaitSystemEventResponse struct {
	PResult     ndr.DWORD
	PEventFlags ndr.DWORD
	Status      ndr.DWORD `ndr:"retval"`
}

// RpcWinStationWaitSystemEvent calls RpcWinStationWaitSystemEvent (opnum 16) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcWinStationWaitSystemEvent(rpc ndr.Invoker, hServer mststs.SERVER_HANDLE, eventMask ndr.DWORD) (PResult ndr.DWORD, PEventFlags ndr.DWORD, err error) {
	req := &rpcWinStationWaitSystemEventRequest{
		HServer:   hServer,
		EventMask: eventMask,
	}
	var resp rpcWinStationWaitSystemEventResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcWinStationWaitSystemEvent: %w", err)
		return
	}
	PResult = resp.PResult
	PEventFlags = resp.PEventFlags
	if uint32(resp.Status) != IcaApi.StatusSuccess {
		err = fmt.Errorf("RpcWinStationWaitSystemEvent failed: %s", IcaApi.StatusString(uint32(resp.Status)))
	}
	return
}
