package functions

import (
	"fmt"

	TSVIPPublic "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/53b46b02-c73b-4a3e-8dee-b16b80672fc0/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mststs "github.com/TheManticoreProject/Manticore/windows/protocols/ms-tsts"
)

// rpcGetSessionIPRequest carries the [in] parameters of RpcGetSessionIP.
type rpcGetSessionIPRequest struct {
	Family    uint16
	SessionId ndr.DWORD
}

func (*rpcGetSessionIPRequest) Opnum() uint16 { return TSVIPPublic.OpnumRpcGetSessionIP }

// rpcGetSessionIPResponse carries the [out] parameters and return value of RpcGetSessionIP.
type rpcGetSessionIPResponse struct {
	PpVIPSession mststs.TSVIPSession
	Status       ndr.DWORD `ndr:"retval"`
}

// RpcGetSessionIP calls RpcGetSessionIP (opnum 0) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcGetSessionIP(rpc ndr.Invoker, family uint16, sessionId ndr.DWORD) (PpVIPSession mststs.TSVIPSession, err error) {
	req := &rpcGetSessionIPRequest{
		Family:    family,
		SessionId: sessionId,
	}
	var resp rpcGetSessionIPResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcGetSessionIP: %w", err)
		return
	}
	PpVIPSession = resp.PpVIPSession
	if uint32(resp.Status) != TSVIPPublic.StatusSuccess {
		err = fmt.Errorf("RpcGetSessionIP failed: %s", TSVIPPublic.StatusString(uint32(resp.Status)))
	}
	return
}
