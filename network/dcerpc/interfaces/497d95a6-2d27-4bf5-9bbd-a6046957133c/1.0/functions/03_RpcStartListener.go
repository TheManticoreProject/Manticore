package functions

import (
	"fmt"

	RCMListener "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/497d95a6-2d27-4bf5-9bbd-a6046957133c/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mststs "github.com/TheManticoreProject/Manticore/windows/protocols/ms-tsts"
)

// rpcStartListenerRequest carries the [in] parameters of RpcStartListener.
type rpcStartListenerRequest struct {
	HListener mststs.HLISTENER
}

func (*rpcStartListenerRequest) Opnum() uint16 { return RCMListener.OpnumRpcStartListener }

// rpcStartListenerResponse carries the [out] parameters and return value of RpcStartListener.
type rpcStartListenerResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RpcStartListener calls RpcStartListener (opnum 3) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcStartListener(rpc ndr.Invoker, hListener mststs.HLISTENER) (err error) {
	req := &rpcStartListenerRequest{
		HListener: hListener,
	}
	var resp rpcStartListenerResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcStartListener: %w", err)
		return
	}
	if uint32(resp.Status) != RCMListener.StatusSuccess {
		err = fmt.Errorf("RpcStartListener failed: %s", RCMListener.StatusString(uint32(resp.Status)))
	}
	return
}
