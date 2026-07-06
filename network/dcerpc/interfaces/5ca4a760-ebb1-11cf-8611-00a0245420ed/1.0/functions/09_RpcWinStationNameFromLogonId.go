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

// rpcWinStationNameFromLogonIdRequest carries the [in] parameters of RpcWinStationNameFromLogonId.
type rpcWinStationNameFromLogonIdRequest struct {
	HServer         mststs.SERVER_HANDLE
	LoginId         ndr.DWORD
	PWinStationName []uint16 `ndr:"ref,size_is=NameSize"`
	NameSize        ndr.DWORD
}

func (*rpcWinStationNameFromLogonIdRequest) Opnum() uint16 {
	return IcaApi.OpnumRpcWinStationNameFromLogonId
}

// rpcWinStationNameFromLogonIdResponse carries the [out] parameters and return value of RpcWinStationNameFromLogonId.
type rpcWinStationNameFromLogonIdResponse struct {
	PResult         ndr.DWORD
	PWinStationName []uint16  `ndr:"ref,size_is=NameSize"`
	Status          ndr.DWORD `ndr:"retval"`
}

// RpcWinStationNameFromLogonId calls RpcWinStationNameFromLogonId (opnum 9) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcWinStationNameFromLogonId(rpc ndr.Invoker, hServer mststs.SERVER_HANDLE, loginId ndr.DWORD, pWinStationName []uint16, nameSize ndr.DWORD) (PResult ndr.DWORD, PWinStationName []uint16, err error) {
	req := &rpcWinStationNameFromLogonIdRequest{
		HServer:         hServer,
		LoginId:         loginId,
		PWinStationName: pWinStationName,
		NameSize:        nameSize,
	}
	var resp rpcWinStationNameFromLogonIdResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcWinStationNameFromLogonId: %w", err)
		return
	}
	PResult = resp.PResult
	PWinStationName = resp.PWinStationName
	if uint32(resp.Status) != IcaApi.StatusSuccess {
		err = fmt.Errorf("RpcWinStationNameFromLogonId failed: %s", IcaApi.StatusString(uint32(resp.Status)))
	}
	return
}
