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

// rpcWinStationGetAllProcesses_NT6Request carries the [in] parameters of RpcWinStationGetAllProcesses_NT6.
type rpcWinStationGetAllProcesses_NT6Request struct {
	HServer            mststs.SERVER_HANDLE
	Level              ndr.DWORD
	PNumberOfProcesses ndr.DWORD
}

func (*rpcWinStationGetAllProcesses_NT6Request) Opnum() uint16 {
	return IcaApi.OpnumRpcWinStationGetAllProcesses_NT6
}

// rpcWinStationGetAllProcesses_NT6Response carries the [out] parameters and return value of RpcWinStationGetAllProcesses_NT6.
type rpcWinStationGetAllProcesses_NT6Response struct {
	PResult              ndr.DWORD
	PNumberOfProcesses   ndr.DWORD
	PpTsAllProcessesInfo []mststs.TS_ALL_PROCESSES_INFO_NT6 `ndr:"unique,conformant"`
	Status               ndr.DWORD                          `ndr:"retval"`
}

// RpcWinStationGetAllProcesses_NT6 calls RpcWinStationGetAllProcesses_NT6 (opnum 70) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcWinStationGetAllProcesses_NT6(rpc ndr.Invoker, hServer mststs.SERVER_HANDLE, level ndr.DWORD, pNumberOfProcesses ndr.DWORD) (PResult ndr.DWORD, PNumberOfProcesses ndr.DWORD, PpTsAllProcessesInfo []mststs.TS_ALL_PROCESSES_INFO_NT6, err error) {
	req := &rpcWinStationGetAllProcesses_NT6Request{
		HServer:            hServer,
		Level:              level,
		PNumberOfProcesses: pNumberOfProcesses,
	}
	var resp rpcWinStationGetAllProcesses_NT6Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcWinStationGetAllProcesses_NT6: %w", err)
		return
	}
	PResult = resp.PResult
	PNumberOfProcesses = resp.PNumberOfProcesses
	PpTsAllProcessesInfo = resp.PpTsAllProcessesInfo
	if uint32(resp.Status) != IcaApi.StatusSuccess {
		err = fmt.Errorf("RpcWinStationGetAllProcesses_NT6 failed: %s", IcaApi.StatusString(uint32(resp.Status)))
	}
	return
}
