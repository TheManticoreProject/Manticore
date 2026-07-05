package functions

import (
	"fmt"

	IcaApi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5ca4a760-ebb1-11cf-8611-00a0245420ed/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mststs "github.com/TheManticoreProject/Manticore/windows/protocols/ms-tsts"
)

// rpcWinStationGetTermSrvCountersValueRequest carries the [in] parameters of RpcWinStationGetTermSrvCountersValue.
type rpcWinStationGetTermSrvCountersValueRequest struct {
	HServer   mststs.SERVER_HANDLE
	DwEntries ndr.DWORD
	PCounter  []mststs.TS_COUNTER `ndr:"ref,size_is=DwEntries"`
}

func (*rpcWinStationGetTermSrvCountersValueRequest) Opnum() uint16 {
	return IcaApi.OpnumRpcWinStationGetTermSrvCountersValue
}

// rpcWinStationGetTermSrvCountersValueResponse carries the [out] parameters and return value of RpcWinStationGetTermSrvCountersValue.
type rpcWinStationGetTermSrvCountersValueResponse struct {
	PResult  ndr.DWORD
	PCounter []mststs.TS_COUNTER `ndr:"ref,size_is=DwEntries"`
	Status   ndr.DWORD           `ndr:"retval"`
}

// RpcWinStationGetTermSrvCountersValue calls RpcWinStationGetTermSrvCountersValue (opnum 45) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcWinStationGetTermSrvCountersValue(rpc ndr.Invoker, hServer mststs.SERVER_HANDLE, dwEntries ndr.DWORD, pCounter []mststs.TS_COUNTER) (PResult ndr.DWORD, PCounter []mststs.TS_COUNTER, err error) {
	req := &rpcWinStationGetTermSrvCountersValueRequest{
		HServer:   hServer,
		DwEntries: dwEntries,
		PCounter:  pCounter,
	}
	var resp rpcWinStationGetTermSrvCountersValueResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcWinStationGetTermSrvCountersValue: %w", err)
		return
	}
	PResult = resp.PResult
	PCounter = resp.PCounter
	if uint32(resp.Status) != IcaApi.StatusSuccess {
		err = fmt.Errorf("RpcWinStationGetTermSrvCountersValue failed: %s", IcaApi.StatusString(uint32(resp.Status)))
	}
	return
}
