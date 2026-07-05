package functions

import (
	"fmt"

	RCMPublic "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/bde95fdf-eee0-45de-9e12-e5a61cd0d4fe/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mststs "github.com/TheManticoreProject/Manticore/windows/protocols/ms-tsts"
)

// rpcGetSessionProtocolLastInputTimeRequest carries the [in] parameters of RpcGetSessionProtocolLastInputTime.
type rpcGetSessionProtocolLastInputTimeRequest struct {
	SessionId ndr.DWORD
	InfoType  mststs.PROTOCOLSTATUS_INFO_TYPE
}

func (*rpcGetSessionProtocolLastInputTimeRequest) Opnum() uint16 {
	return RCMPublic.OpnumRpcGetSessionProtocolLastInputTime
}

// rpcGetSessionProtocolLastInputTimeResponse carries the [out] parameters and return value of RpcGetSessionProtocolLastInputTime.
type rpcGetSessionProtocolLastInputTimeResponse struct {
	PpProtoStatus  []uint8 `ndr:"unique,conformant"`
	PcbProtoStatus ndr.DWORD
	PLastInputTime int64
	Status         ndr.DWORD `ndr:"retval"`
}

// RpcGetSessionProtocolLastInputTime calls RpcGetSessionProtocolLastInputTime (opnum 9) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcGetSessionProtocolLastInputTime(rpc ndr.Invoker, sessionId ndr.DWORD, infoType mststs.PROTOCOLSTATUS_INFO_TYPE) (PpProtoStatus []uint8, PcbProtoStatus ndr.DWORD, PLastInputTime int64, err error) {
	req := &rpcGetSessionProtocolLastInputTimeRequest{
		SessionId: sessionId,
		InfoType:  infoType,
	}
	var resp rpcGetSessionProtocolLastInputTimeResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcGetSessionProtocolLastInputTime: %w", err)
		return
	}
	PpProtoStatus = resp.PpProtoStatus
	PcbProtoStatus = resp.PcbProtoStatus
	PLastInputTime = resp.PLastInputTime
	if uint32(resp.Status) != RCMPublic.StatusSuccess {
		err = fmt.Errorf("RpcGetSessionProtocolLastInputTime failed: %s", RCMPublic.StatusString(uint32(resp.Status)))
	}
	return
}
