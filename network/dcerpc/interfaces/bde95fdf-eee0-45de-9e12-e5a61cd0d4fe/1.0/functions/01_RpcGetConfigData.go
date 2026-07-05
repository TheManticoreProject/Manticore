package functions

import (
	"fmt"

	RCMPublic "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/bde95fdf-eee0-45de-9e12-e5a61cd0d4fe/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rpcGetConfigDataRequest carries the [in] parameters of RpcGetConfigData.
type rpcGetConfigDataRequest struct {
	SessionId ndr.DWORD
}

func (*rpcGetConfigDataRequest) Opnum() uint16 { return RCMPublic.OpnumRpcGetConfigData }

// rpcGetConfigDataResponse carries the [out] parameters and return value of RpcGetConfigData.
type rpcGetConfigDataResponse struct {
	PpBuff          []uint8 `ndr:"unique,conformant"`
	POutBuffByteLen ndr.DWORD
	Status          ndr.DWORD `ndr:"retval"`
}

// RpcGetConfigData calls RpcGetConfigData (opnum 1) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcGetConfigData(rpc ndr.Invoker, sessionId ndr.DWORD) (PpBuff []uint8, POutBuffByteLen ndr.DWORD, err error) {
	req := &rpcGetConfigDataRequest{
		SessionId: sessionId,
	}
	var resp rpcGetConfigDataResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcGetConfigData: %w", err)
		return
	}
	PpBuff = resp.PpBuff
	POutBuffByteLen = resp.POutBuffByteLen
	if uint32(resp.Status) != RCMPublic.StatusSuccess {
		err = fmt.Errorf("RpcGetConfigData failed: %s", RCMPublic.StatusString(uint32(resp.Status)))
	}
	return
}
