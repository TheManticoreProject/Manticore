package functions

import (
	"fmt"

	TermSrvSession "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/484809d6-4239-471b-b5bc-61df8c23ac48/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mststs "github.com/TheManticoreProject/Manticore/windows/protocols/ms-tsts"
)

// rpcGetSessionInformationExRequest carries the [in] parameters of RpcGetSessionInformationEx.
type rpcGetSessionInformationExRequest struct {
	SessionId int32
	Level     ndr.DWORD
}

func (*rpcGetSessionInformationExRequest) Opnum() uint16 {
	return TermSrvSession.OpnumRpcGetSessionInformationEx
}

// rpcGetSessionInformationExResponse carries the [out] parameters and return value of RpcGetSessionInformationEx.
type rpcGetSessionInformationExResponse struct {
	LSMSessionInfoExPtr mststs.LSMSESSIONINFORMATION_EX
	Status              ndr.DWORD `ndr:"retval"`
}

// RpcGetSessionInformationEx calls RpcGetSessionInformationEx (opnum 17) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcGetSessionInformationEx(rpc ndr.Invoker, sessionId int32, level ndr.DWORD) (LSMSessionInfoExPtr mststs.LSMSESSIONINFORMATION_EX, err error) {
	req := &rpcGetSessionInformationExRequest{
		SessionId: sessionId,
		Level:     level,
	}
	var resp rpcGetSessionInformationExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcGetSessionInformationEx: %w", err)
		return
	}
	LSMSessionInfoExPtr = resp.LSMSessionInfoExPtr
	if uint32(resp.Status) != TermSrvSession.StatusSuccess {
		err = fmt.Errorf("RpcGetSessionInformationEx failed: %s", TermSrvSession.StatusString(uint32(resp.Status)))
	}
	return
}
