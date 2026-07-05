package functions

import (
	"fmt"

	IcaApi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5ca4a760-ebb1-11cf-8611-00a0245420ed/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mststs "github.com/TheManticoreProject/Manticore/windows/protocols/ms-tsts"
)

// rpcWinStationConnectRequest carries the [in] parameters of RpcWinStationConnect.
type rpcWinStationConnectRequest struct {
	HServer        mststs.SERVER_HANDLE
	ClientLogonId  ndr.DWORD
	ConnectLogonId ndr.DWORD
	TargetLogonId  ndr.DWORD
	PPassword      []uint16 `ndr:"ref,size_is=PasswordSize"`
	PasswordSize   ndr.DWORD
	Wait           bool
}

func (*rpcWinStationConnectRequest) Opnum() uint16 { return IcaApi.OpnumRpcWinStationConnect }

// rpcWinStationConnectResponse carries the [out] parameters and return value of RpcWinStationConnect.
type rpcWinStationConnectResponse struct {
	PResult ndr.DWORD
	Status  ndr.DWORD `ndr:"retval"`
}

// RpcWinStationConnect calls RpcWinStationConnect (opnum 10) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcWinStationConnect(rpc ndr.Invoker, hServer mststs.SERVER_HANDLE, clientLogonId ndr.DWORD, connectLogonId ndr.DWORD, targetLogonId ndr.DWORD, pPassword []uint16, passwordSize ndr.DWORD, wait bool) (PResult ndr.DWORD, err error) {
	req := &rpcWinStationConnectRequest{
		HServer:        hServer,
		ClientLogonId:  clientLogonId,
		ConnectLogonId: connectLogonId,
		TargetLogonId:  targetLogonId,
		PPassword:      pPassword,
		PasswordSize:   passwordSize,
		Wait:           wait,
	}
	var resp rpcWinStationConnectResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcWinStationConnect: %w", err)
		return
	}
	PResult = resp.PResult
	if uint32(resp.Status) != IcaApi.StatusSuccess {
		err = fmt.Errorf("RpcWinStationConnect failed: %s", IcaApi.StatusString(uint32(resp.Status)))
	}
	return
}
