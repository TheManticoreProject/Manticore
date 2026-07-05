package functions

import (
	"fmt"

	IcaApi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5ca4a760-ebb1-11cf-8611-00a0245420ed/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mststs "github.com/TheManticoreProject/Manticore/windows/protocols/ms-tsts"
)

// rpcWinStationEnumerateRequest carries the [in] parameters of RpcWinStationEnumerate.
type rpcWinStationEnumerateRequest struct {
	HServer    mststs.SERVER_HANDLE
	PEntries   ndr.DWORD
	PLogonId   []int8 `ndr:"ref,conformant"`
	PByteCount ndr.DWORD
	PIndex     ndr.DWORD
}

func (*rpcWinStationEnumerateRequest) Opnum() uint16 { return IcaApi.OpnumRpcWinStationEnumerate }

// rpcWinStationEnumerateResponse carries the [out] parameters and return value of RpcWinStationEnumerate.
type rpcWinStationEnumerateResponse struct {
	PResult    ndr.DWORD
	PEntries   ndr.DWORD
	PLogonId   []int8 `ndr:"ref,conformant"`
	PByteCount ndr.DWORD
	PIndex     ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// RpcWinStationEnumerate calls RpcWinStationEnumerate (opnum 3) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcWinStationEnumerate(rpc ndr.Invoker, hServer mststs.SERVER_HANDLE, pEntries ndr.DWORD, pLogonId []int8, pByteCount ndr.DWORD, pIndex ndr.DWORD) (PResult ndr.DWORD, PEntries ndr.DWORD, PLogonId []int8, PByteCount ndr.DWORD, PIndex ndr.DWORD, err error) {
	req := &rpcWinStationEnumerateRequest{
		HServer:    hServer,
		PEntries:   pEntries,
		PLogonId:   pLogonId,
		PByteCount: pByteCount,
		PIndex:     pIndex,
	}
	var resp rpcWinStationEnumerateResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcWinStationEnumerate: %w", err)
		return
	}
	PResult = resp.PResult
	PEntries = resp.PEntries
	PLogonId = resp.PLogonId
	PByteCount = resp.PByteCount
	PIndex = resp.PIndex
	if uint32(resp.Status) != IcaApi.StatusSuccess {
		err = fmt.Errorf("RpcWinStationEnumerate failed: %s", IcaApi.StatusString(uint32(resp.Status)))
	}
	return
}
