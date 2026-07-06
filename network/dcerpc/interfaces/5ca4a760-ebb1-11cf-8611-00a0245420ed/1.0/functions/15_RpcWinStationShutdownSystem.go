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

// rpcWinStationShutdownSystemRequest carries the [in] parameters of RpcWinStationShutdownSystem.
type rpcWinStationShutdownSystemRequest struct {
	HServer       mststs.SERVER_HANDLE
	ClientLogonId ndr.DWORD
	ShutdownFlags ndr.DWORD
}

func (*rpcWinStationShutdownSystemRequest) Opnum() uint16 {
	return IcaApi.OpnumRpcWinStationShutdownSystem
}

// rpcWinStationShutdownSystemResponse carries the [out] parameters and return value of RpcWinStationShutdownSystem.
type rpcWinStationShutdownSystemResponse struct {
	PResult ndr.DWORD
	Status  ndr.DWORD `ndr:"retval"`
}

// RpcWinStationShutdownSystem calls RpcWinStationShutdownSystem (opnum 15) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcWinStationShutdownSystem(rpc ndr.Invoker, hServer mststs.SERVER_HANDLE, clientLogonId ndr.DWORD, shutdownFlags ndr.DWORD) (PResult ndr.DWORD, err error) {
	req := &rpcWinStationShutdownSystemRequest{
		HServer:       hServer,
		ClientLogonId: clientLogonId,
		ShutdownFlags: shutdownFlags,
	}
	var resp rpcWinStationShutdownSystemResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcWinStationShutdownSystem: %w", err)
		return
	}
	PResult = resp.PResult
	if uint32(resp.Status) != IcaApi.StatusSuccess {
		err = fmt.Errorf("RpcWinStationShutdownSystem failed: %s", IcaApi.StatusString(uint32(resp.Status)))
	}
	return
}
