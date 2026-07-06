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

// rpcWinStationCloseServerExRequest carries the [in] parameters of RpcWinStationCloseServerEx.
type rpcWinStationCloseServerExRequest struct {
	PhServer mststs.SERVER_HANDLE
}

func (*rpcWinStationCloseServerExRequest) Opnum() uint16 {
	return IcaApi.OpnumRpcWinStationCloseServerEx
}

// rpcWinStationCloseServerExResponse carries the [out] parameters and return value of RpcWinStationCloseServerEx.
type rpcWinStationCloseServerExResponse struct {
	PhServer mststs.SERVER_HANDLE
	PResult  ndr.DWORD
	Status   ndr.DWORD `ndr:"retval"`
}

// RpcWinStationCloseServerEx calls RpcWinStationCloseServerEx (opnum 60) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcWinStationCloseServerEx(rpc ndr.Invoker, phServer mststs.SERVER_HANDLE) (PhServer mststs.SERVER_HANDLE, PResult ndr.DWORD, err error) {
	req := &rpcWinStationCloseServerExRequest{
		PhServer: phServer,
	}
	var resp rpcWinStationCloseServerExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcWinStationCloseServerEx: %w", err)
		return
	}
	PhServer = resp.PhServer
	PResult = resp.PResult
	if uint32(resp.Status) != IcaApi.StatusSuccess {
		err = fmt.Errorf("RpcWinStationCloseServerEx failed: %s", IcaApi.StatusString(uint32(resp.Status)))
	}
	return
}
