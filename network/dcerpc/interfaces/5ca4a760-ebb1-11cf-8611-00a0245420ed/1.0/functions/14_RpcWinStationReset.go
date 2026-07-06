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

// rpcWinStationResetRequest carries the [in] parameters of RpcWinStationReset.
type rpcWinStationResetRequest struct {
	HServer mststs.SERVER_HANDLE
	LogonId ndr.DWORD
	BWait   bool
}

func (*rpcWinStationResetRequest) Opnum() uint16 { return IcaApi.OpnumRpcWinStationReset }

// rpcWinStationResetResponse carries the [out] parameters and return value of RpcWinStationReset.
type rpcWinStationResetResponse struct {
	PResult ndr.DWORD
	Status  ndr.DWORD `ndr:"retval"`
}

// RpcWinStationReset calls RpcWinStationReset (opnum 14) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcWinStationReset(rpc ndr.Invoker, hServer mststs.SERVER_HANDLE, logonId ndr.DWORD, bWait bool) (PResult ndr.DWORD, err error) {
	req := &rpcWinStationResetRequest{
		HServer: hServer,
		LogonId: logonId,
		BWait:   bWait,
	}
	var resp rpcWinStationResetResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcWinStationReset: %w", err)
		return
	}
	PResult = resp.PResult
	if uint32(resp.Status) != IcaApi.StatusSuccess {
		err = fmt.Errorf("RpcWinStationReset failed: %s", IcaApi.StatusString(uint32(resp.Status)))
	}
	return
}
