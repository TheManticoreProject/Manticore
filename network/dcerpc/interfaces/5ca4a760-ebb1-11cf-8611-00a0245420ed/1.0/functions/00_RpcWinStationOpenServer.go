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

// rpcWinStationOpenServerRequest carries the [in] parameters of RpcWinStationOpenServer.
type rpcWinStationOpenServerRequest struct {
}

func (*rpcWinStationOpenServerRequest) Opnum() uint16 { return IcaApi.OpnumRpcWinStationOpenServer }

// rpcWinStationOpenServerResponse carries the [out] parameters and return value of RpcWinStationOpenServer.
type rpcWinStationOpenServerResponse struct {
	PResult  ndr.DWORD
	PhServer mststs.SERVER_HANDLE
	Status   ndr.DWORD `ndr:"retval"`
}

// RpcWinStationOpenServer calls RpcWinStationOpenServer (opnum 0) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcWinStationOpenServer(rpc ndr.Invoker) (PResult ndr.DWORD, PhServer mststs.SERVER_HANDLE, err error) {
	req := &rpcWinStationOpenServerRequest{}
	var resp rpcWinStationOpenServerResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcWinStationOpenServer: %w", err)
		return
	}
	PResult = resp.PResult
	PhServer = resp.PhServer
	if uint32(resp.Status) != IcaApi.StatusSuccess {
		err = fmt.Errorf("RpcWinStationOpenServer failed: %s", IcaApi.StatusString(uint32(resp.Status)))
	}
	return
}
