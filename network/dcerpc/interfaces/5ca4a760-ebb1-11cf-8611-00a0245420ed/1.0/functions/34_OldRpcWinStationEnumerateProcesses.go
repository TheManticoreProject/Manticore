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

// oldRpcWinStationEnumerateProcessesRequest carries the [in] parameters of OldRpcWinStationEnumerateProcesses.
type oldRpcWinStationEnumerateProcessesRequest struct {
	HServer   mststs.SERVER_HANDLE
	ByteCount ndr.DWORD
}

func (*oldRpcWinStationEnumerateProcessesRequest) Opnum() uint16 {
	return IcaApi.OpnumOldRpcWinStationEnumerateProcesses
}

// oldRpcWinStationEnumerateProcessesResponse carries the [out] parameters and return value of OldRpcWinStationEnumerateProcesses.
type oldRpcWinStationEnumerateProcessesResponse struct {
	PResult        ndr.DWORD
	PProcessBuffer []uint8   `ndr:"ref,size_is=ByteCount"`
	Status         ndr.DWORD `ndr:"retval"`
}

// OldRpcWinStationEnumerateProcesses calls OldRpcWinStationEnumerateProcesses (opnum 34) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func OldRpcWinStationEnumerateProcesses(rpc ndr.Invoker, hServer mststs.SERVER_HANDLE, byteCount ndr.DWORD) (PResult ndr.DWORD, PProcessBuffer []uint8, err error) {
	req := &oldRpcWinStationEnumerateProcessesRequest{
		HServer:   hServer,
		ByteCount: byteCount,
	}
	var resp oldRpcWinStationEnumerateProcessesResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("OldRpcWinStationEnumerateProcesses: %w", err)
		return
	}
	PResult = resp.PResult
	PProcessBuffer = resp.PProcessBuffer
	if uint32(resp.Status) != IcaApi.StatusSuccess {
		err = fmt.Errorf("OldRpcWinStationEnumerateProcesses failed: %s", IcaApi.StatusString(uint32(resp.Status)))
	}
	return
}
