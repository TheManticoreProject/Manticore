package functions

import (
	"fmt"

	RCMListener "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/497d95a6-2d27-4bf5-9bbd-a6046957133c/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mststs "github.com/TheManticoreProject/Manticore/windows/protocols/ms-tsts"
)

// rpcIsListeningRequest carries the [in] parameters of RpcIsListening.
type rpcIsListeningRequest struct {
	HListener mststs.HLISTENER
}

func (*rpcIsListeningRequest) Opnum() uint16 { return RCMListener.OpnumRpcIsListening }

// rpcIsListeningResponse carries the [out] parameters and return value of RpcIsListening.
type rpcIsListeningResponse struct {
	PbIsListening ndr.BOOL
	Status        ndr.DWORD `ndr:"retval"`
}

// RpcIsListening calls RpcIsListening (opnum 4) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcIsListening(rpc ndr.Invoker, hListener mststs.HLISTENER) (PbIsListening ndr.BOOL, err error) {
	req := &rpcIsListeningRequest{
		HListener: hListener,
	}
	var resp rpcIsListeningResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcIsListening: %w", err)
		return
	}
	PbIsListening = resp.PbIsListening
	if uint32(resp.Status) != RCMListener.StatusSuccess {
		err = fmt.Errorf("RpcIsListening failed: %s", RCMListener.StatusString(uint32(resp.Status)))
	}
	return
}
