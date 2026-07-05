package functions

import (
	"fmt"

	IcaApi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5ca4a760-ebb1-11cf-8611-00a0245420ed/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mststs "github.com/TheManticoreProject/Manticore/windows/protocols/ms-tsts"
)

// rpcWinStationQueryInformationRequest carries the [in] parameters of RpcWinStationQueryInformation.
type rpcWinStationQueryInformationRequest struct {
	HServer                     mststs.SERVER_HANDLE
	LogonId                     ndr.DWORD
	WinStationInformationClass  ndr.DWORD
	PWinStationInformation      []int8 `ndr:"ref,size_is=WinStationInformationLength"`
	WinStationInformationLength ndr.DWORD
}

func (*rpcWinStationQueryInformationRequest) Opnum() uint16 {
	return IcaApi.OpnumRpcWinStationQueryInformation
}

// rpcWinStationQueryInformationResponse carries the [out] parameters and return value of RpcWinStationQueryInformation.
type rpcWinStationQueryInformationResponse struct {
	PResult                ndr.DWORD
	PWinStationInformation []int8 `ndr:"ref,size_is=WinStationInformationLength"`
	PReturnLength          ndr.DWORD
	Status                 ndr.DWORD `ndr:"retval"`
}

// RpcWinStationQueryInformation calls RpcWinStationQueryInformation (opnum 5) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcWinStationQueryInformation(rpc ndr.Invoker, hServer mststs.SERVER_HANDLE, logonId ndr.DWORD, winStationInformationClass ndr.DWORD, pWinStationInformation []int8, winStationInformationLength ndr.DWORD) (PResult ndr.DWORD, PWinStationInformation []int8, PReturnLength ndr.DWORD, err error) {
	req := &rpcWinStationQueryInformationRequest{
		HServer:                     hServer,
		LogonId:                     logonId,
		WinStationInformationClass:  winStationInformationClass,
		PWinStationInformation:      pWinStationInformation,
		WinStationInformationLength: winStationInformationLength,
	}
	var resp rpcWinStationQueryInformationResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcWinStationQueryInformation: %w", err)
		return
	}
	PResult = resp.PResult
	PWinStationInformation = resp.PWinStationInformation
	PReturnLength = resp.PReturnLength
	if uint32(resp.Status) != IcaApi.StatusSuccess {
		err = fmt.Errorf("RpcWinStationQueryInformation failed: %s", IcaApi.StatusString(uint32(resp.Status)))
	}
	return
}
