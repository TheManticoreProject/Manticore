package functions

import (
	"fmt"

	RCMPublic "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/bde95fdf-eee0-45de-9e12-e5a61cd0d4fe/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rpcGetClientDataRequest carries the [in] parameters of RpcGetClientData.
type rpcGetClientDataRequest struct {
	SessionId ndr.DWORD
}

func (*rpcGetClientDataRequest) Opnum() uint16 { return RCMPublic.OpnumRpcGetClientData }

// rpcGetClientDataResponse carries the [out] parameters and return value of RpcGetClientData.
type rpcGetClientDataResponse struct {
	PpBuff          []uint8 `ndr:"unique,conformant"`
	POutBuffByteLen ndr.DWORD
	Status          ndr.DWORD `ndr:"retval"`
}

// RpcGetClientData calls RpcGetClientData (opnum 0) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcGetClientData(rpc ndr.Invoker, sessionId ndr.DWORD) (PpBuff []uint8, POutBuffByteLen ndr.DWORD, err error) {
	req := &rpcGetClientDataRequest{
		SessionId: sessionId,
	}
	var resp rpcGetClientDataResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcGetClientData: %w", err)
		return
	}
	PpBuff = resp.PpBuff
	POutBuffByteLen = resp.POutBuffByteLen
	if uint32(resp.Status) != RCMPublic.StatusSuccess {
		err = fmt.Errorf("RpcGetClientData failed: %s", RCMPublic.StatusString(uint32(resp.Status)))
	}
	return
}
