package functions

import (
	"fmt"

	IcaApi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5ca4a760-ebb1-11cf-8611-00a0245420ed/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
	mststs "github.com/TheManticoreProject/Manticore/windows/protocols/ms-tsts"
)

// rpcWinStationGetProcessSidRequest carries the [in] parameters of RpcWinStationGetProcessSid.
type rpcWinStationGetProcessSidRequest struct {
	HServer           mststs.SERVER_HANDLE
	DwUniqueProcessId ndr.DWORD
	ProcessStartTime  msdtyp.LARGE_INTEGER
	PProcessUserSid   []uint8 `ndr:"ref,size_is=DwSidSize"`
	DwSidSize         ndr.DWORD
	PdwSizeNeeded     ndr.DWORD
}

func (*rpcWinStationGetProcessSidRequest) Opnum() uint16 {
	return IcaApi.OpnumRpcWinStationGetProcessSid
}

// rpcWinStationGetProcessSidResponse carries the [out] parameters and return value of RpcWinStationGetProcessSid.
type rpcWinStationGetProcessSidResponse struct {
	PResult         int32
	PProcessUserSid []uint8 `ndr:"ref,size_is=DwSidSize"`
	PdwSizeNeeded   ndr.DWORD
	Status          ndr.DWORD `ndr:"retval"`
}

// RpcWinStationGetProcessSid calls RpcWinStationGetProcessSid (opnum 44) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcWinStationGetProcessSid(rpc ndr.Invoker, hServer mststs.SERVER_HANDLE, dwUniqueProcessId ndr.DWORD, processStartTime msdtyp.LARGE_INTEGER, pProcessUserSid []uint8, dwSidSize ndr.DWORD, pdwSizeNeeded ndr.DWORD) (PResult int32, PProcessUserSid []uint8, PdwSizeNeeded ndr.DWORD, err error) {
	req := &rpcWinStationGetProcessSidRequest{
		HServer:           hServer,
		DwUniqueProcessId: dwUniqueProcessId,
		ProcessStartTime:  processStartTime,
		PProcessUserSid:   pProcessUserSid,
		DwSidSize:         dwSidSize,
		PdwSizeNeeded:     pdwSizeNeeded,
	}
	var resp rpcWinStationGetProcessSidResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcWinStationGetProcessSid: %w", err)
		return
	}
	PResult = resp.PResult
	PProcessUserSid = resp.PProcessUserSid
	PdwSizeNeeded = resp.PdwSizeNeeded
	if uint32(resp.Status) != IcaApi.StatusSuccess {
		err = fmt.Errorf("RpcWinStationGetProcessSid failed: %s", IcaApi.StatusString(uint32(resp.Status)))
	}
	return
}
