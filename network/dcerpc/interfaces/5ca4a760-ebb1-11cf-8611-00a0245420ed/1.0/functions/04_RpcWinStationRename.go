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

// rpcWinStationRenameRequest carries the [in] parameters of RpcWinStationRename.
type rpcWinStationRenameRequest struct {
	HServer            mststs.SERVER_HANDLE
	PWinStationNameOld []uint16 `ndr:"ref,size_is=NameOldSize"`
	NameOldSize        ndr.DWORD
	PWinStationNameNew []uint16 `ndr:"ref,size_is=NameNewSize"`
	NameNewSize        ndr.DWORD
}

func (*rpcWinStationRenameRequest) Opnum() uint16 { return IcaApi.OpnumRpcWinStationRename }

// rpcWinStationRenameResponse carries the [out] parameters and return value of RpcWinStationRename.
type rpcWinStationRenameResponse struct {
	PResult ndr.DWORD
	Status  ndr.DWORD `ndr:"retval"`
}

// RpcWinStationRename calls RpcWinStationRename (opnum 4) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcWinStationRename(rpc ndr.Invoker, hServer mststs.SERVER_HANDLE, pWinStationNameOld []uint16, nameOldSize ndr.DWORD, pWinStationNameNew []uint16, nameNewSize ndr.DWORD) (PResult ndr.DWORD, err error) {
	req := &rpcWinStationRenameRequest{
		HServer:            hServer,
		PWinStationNameOld: pWinStationNameOld,
		NameOldSize:        nameOldSize,
		PWinStationNameNew: pWinStationNameNew,
		NameNewSize:        nameNewSize,
	}
	var resp rpcWinStationRenameResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcWinStationRename: %w", err)
		return
	}
	PResult = resp.PResult
	if uint32(resp.Status) != IcaApi.StatusSuccess {
		err = fmt.Errorf("RpcWinStationRename failed: %s", IcaApi.StatusString(uint32(resp.Status)))
	}
	return
}
