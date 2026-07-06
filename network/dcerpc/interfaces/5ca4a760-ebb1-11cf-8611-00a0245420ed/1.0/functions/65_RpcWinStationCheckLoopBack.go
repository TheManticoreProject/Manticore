package functions

// IDL source: [MS-TSTS] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-tsts/c43addc7-eebc-491b-9b01-2587262675e8
// A fetched copy is kept at ms-tsts.idl in the interface directory.

import (
	"fmt"

	IcaApi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5ca4a760-ebb1-11cf-8611-00a0245420ed/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mststs "github.com/TheManticoreProject/Manticore/windows/protocols/ms-tsts"
)

// rpcWinStationCheckLoopBackRequest carries the [in] parameters of RpcWinStationCheckLoopBack.
type rpcWinStationCheckLoopBackRequest struct {
	HServer           mststs.SERVER_HANDLE
	ClientLogonId     ndr.DWORD
	TargetLogonId     ndr.DWORD
	PTargetServerName []uint16 `ndr:"ref,size_is=NameSize"`
	NameSize          ndr.DWORD
}

func (*rpcWinStationCheckLoopBackRequest) Opnum() uint16 {
	return IcaApi.OpnumRpcWinStationCheckLoopBack
}

// rpcWinStationCheckLoopBackResponse carries the [out] parameters and return value of RpcWinStationCheckLoopBack.
type rpcWinStationCheckLoopBackResponse struct {
	PResult ndr.DWORD
	Status  ndr.DWORD `ndr:"retval"`
}

// RpcWinStationCheckLoopBack calls RpcWinStationCheckLoopBack (opnum 65) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcWinStationCheckLoopBack(rpc ndr.Invoker, hServer mststs.SERVER_HANDLE, clientLogonId ndr.DWORD, targetLogonId ndr.DWORD, pTargetServerName []uint16, nameSize ndr.DWORD) (PResult ndr.DWORD, err error) {
	req := &rpcWinStationCheckLoopBackRequest{
		HServer:           hServer,
		ClientLogonId:     clientLogonId,
		TargetLogonId:     targetLogonId,
		PTargetServerName: pTargetServerName,
		NameSize:          nameSize,
	}
	var resp rpcWinStationCheckLoopBackResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcWinStationCheckLoopBack: %w", err)
		return
	}
	PResult = resp.PResult
	if uint32(resp.Status) != IcaApi.StatusSuccess {
		err = fmt.Errorf("RpcWinStationCheckLoopBack failed: %s", IcaApi.StatusString(uint32(resp.Status)))
	}
	return
}
