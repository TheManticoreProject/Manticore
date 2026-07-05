package functions

import (
	"fmt"

	RCMListener "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/497d95a6-2d27-4bf5-9bbd-a6046957133c/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mststs "github.com/TheManticoreProject/Manticore/windows/protocols/ms-tsts"
)

// rpcStopListenerRequest carries the [in] parameters of RpcStopListener.
type rpcStopListenerRequest struct {
	HListener mststs.HLISTENER
}

func (*rpcStopListenerRequest) Opnum() uint16 { return RCMListener.OpnumRpcStopListener }

// rpcStopListenerResponse carries the [out] parameters and return value of RpcStopListener.
type rpcStopListenerResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcStopListener calls RpcStopListener (opnum 2) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcStopListener(rpc ndr.Invoker, hListener mststs.HLISTENER) (err error) {
	req := &rpcStopListenerRequest{
		HListener: hListener,
	}
	var resp rpcStopListenerResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcStopListener: %w", err)
		return
	}
	if uint32(resp.Status) != RCMListener.StatusSuccess {
		err = fmt.Errorf("RpcStopListener failed: %s", RCMListener.StatusString(uint32(resp.Status)))
	}
	return
}
