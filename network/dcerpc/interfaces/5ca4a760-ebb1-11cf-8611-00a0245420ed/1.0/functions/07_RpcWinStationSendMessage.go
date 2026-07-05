package functions

import (
	"fmt"

	IcaApi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5ca4a760-ebb1-11cf-8611-00a0245420ed/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mststs "github.com/TheManticoreProject/Manticore/windows/protocols/ms-tsts"
)

// rpcWinStationSendMessageRequest carries the [in] parameters of RpcWinStationSendMessage.
type rpcWinStationSendMessageRequest struct {
	HServer       mststs.SERVER_HANDLE
	LogonId       ndr.DWORD
	PTitle        []uint16 `ndr:"ref,size_is=TitleLength"`
	TitleLength   ndr.DWORD
	PMessage      []uint16 `ndr:"ref,size_is=MessageLength"`
	MessageLength ndr.DWORD
	Style         ndr.DWORD
	Timeout       ndr.DWORD
	DoNotWait     bool
}

func (*rpcWinStationSendMessageRequest) Opnum() uint16 { return IcaApi.OpnumRpcWinStationSendMessage }

// rpcWinStationSendMessageResponse carries the [out] parameters and return value of RpcWinStationSendMessage.
type rpcWinStationSendMessageResponse struct {
	PResult   ndr.DWORD
	PResponse ndr.DWORD
	Status    ndr.DWORD `ndr:"retval"`
}

// RpcWinStationSendMessage calls RpcWinStationSendMessage (opnum 7) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcWinStationSendMessage(rpc ndr.Invoker, hServer mststs.SERVER_HANDLE, logonId ndr.DWORD, pTitle []uint16, titleLength ndr.DWORD, pMessage []uint16, messageLength ndr.DWORD, style ndr.DWORD, timeout ndr.DWORD, doNotWait bool) (PResult ndr.DWORD, PResponse ndr.DWORD, err error) {
	req := &rpcWinStationSendMessageRequest{
		HServer:       hServer,
		LogonId:       logonId,
		PTitle:        pTitle,
		TitleLength:   titleLength,
		PMessage:      pMessage,
		MessageLength: messageLength,
		Style:         style,
		Timeout:       timeout,
		DoNotWait:     doNotWait,
	}
	var resp rpcWinStationSendMessageResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcWinStationSendMessage: %w", err)
		return
	}
	PResult = resp.PResult
	PResponse = resp.PResponse
	if uint32(resp.Status) != IcaApi.StatusSuccess {
		err = fmt.Errorf("RpcWinStationSendMessage failed: %s", IcaApi.StatusString(uint32(resp.Status)))
	}
	return
}
