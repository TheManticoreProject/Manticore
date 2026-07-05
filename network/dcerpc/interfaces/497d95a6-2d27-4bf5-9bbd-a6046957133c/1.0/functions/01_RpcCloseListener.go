package functions

import (
	"fmt"

	RCMListener "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/497d95a6-2d27-4bf5-9bbd-a6046957133c/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mststs "github.com/TheManticoreProject/Manticore/windows/protocols/ms-tsts"
)

// rpcCloseListenerRequest carries the [in] parameters of RpcCloseListener.
type rpcCloseListenerRequest struct {
	PhListener mststs.HLISTENER
}

func (*rpcCloseListenerRequest) Opnum() uint16 { return RCMListener.OpnumRpcCloseListener }

// rpcCloseListenerResponse carries the [out] parameters and return value of RpcCloseListener.
type rpcCloseListenerResponse struct {
	PhListener mststs.HLISTENER
	Status     ndr.DWORD `ndr:"retval"`
}

// RpcCloseListener calls RpcCloseListener (opnum 1) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcCloseListener(rpc ndr.Invoker, phListener mststs.HLISTENER) (PhListener mststs.HLISTENER, err error) {
	req := &rpcCloseListenerRequest{
		PhListener: phListener,
	}
	var resp rpcCloseListenerResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcCloseListener: %w", err)
		return
	}
	PhListener = resp.PhListener
	if uint32(resp.Status) != RCMListener.StatusSuccess {
		err = fmt.Errorf("RpcCloseListener failed: %s", RCMListener.StatusString(uint32(resp.Status)))
	}
	return
}
