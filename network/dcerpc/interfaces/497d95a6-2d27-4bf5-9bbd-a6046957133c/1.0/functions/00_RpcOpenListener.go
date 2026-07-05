package functions

import (
	"fmt"

	RCMListener "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/497d95a6-2d27-4bf5-9bbd-a6046957133c/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mststs "github.com/TheManticoreProject/Manticore/windows/protocols/ms-tsts"
)

// rpcOpenListenerRequest carries the [in] parameters of RpcOpenListener.
type rpcOpenListenerRequest struct {
	SzListenerName ndr.WSTR
}

func (*rpcOpenListenerRequest) Opnum() uint16 { return RCMListener.OpnumRpcOpenListener }

// rpcOpenListenerResponse carries the [out] parameters and return value of RpcOpenListener.
type rpcOpenListenerResponse struct {
	PhListener mststs.HLISTENER
	Status     ndr.DWORD `ndr:"retval"`
}

// RpcOpenListener calls RpcOpenListener (opnum 0) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcOpenListener(rpc ndr.Invoker, szListenerName ndr.WSTR) (PhListener mststs.HLISTENER, err error) {
	req := &rpcOpenListenerRequest{
		SzListenerName: szListenerName,
	}
	var resp rpcOpenListenerResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcOpenListener: %w", err)
		return
	}
	PhListener = resp.PhListener
	if uint32(resp.Status) != RCMListener.StatusSuccess {
		err = fmt.Errorf("RpcOpenListener failed: %s", RCMListener.StatusString(uint32(resp.Status)))
	}
	return
}
