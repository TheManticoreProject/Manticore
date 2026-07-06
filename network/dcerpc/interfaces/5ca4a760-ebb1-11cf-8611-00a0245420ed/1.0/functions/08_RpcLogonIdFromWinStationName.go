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

// rpcLogonIdFromWinStationNameRequest carries the [in] parameters of RpcLogonIdFromWinStationName.
type rpcLogonIdFromWinStationNameRequest struct {
	HServer         mststs.SERVER_HANDLE
	PWinStationName []uint16 `ndr:"ref,size_is=NameSize"`
	NameSize        ndr.DWORD
}

func (*rpcLogonIdFromWinStationNameRequest) Opnum() uint16 {
	return IcaApi.OpnumRpcLogonIdFromWinStationName
}

// rpcLogonIdFromWinStationNameResponse carries the [out] parameters and return value of RpcLogonIdFromWinStationName.
type rpcLogonIdFromWinStationNameResponse struct {
	PResult  ndr.DWORD
	PLogonId ndr.DWORD
	Status   ndr.DWORD `ndr:"retval"`
}

// RpcLogonIdFromWinStationName calls RpcLogonIdFromWinStationName (opnum 8) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcLogonIdFromWinStationName(rpc ndr.Invoker, hServer mststs.SERVER_HANDLE, pWinStationName []uint16, nameSize ndr.DWORD) (PResult ndr.DWORD, PLogonId ndr.DWORD, err error) {
	req := &rpcLogonIdFromWinStationNameRequest{
		HServer:         hServer,
		PWinStationName: pWinStationName,
		NameSize:        nameSize,
	}
	var resp rpcLogonIdFromWinStationNameResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcLogonIdFromWinStationName: %w", err)
		return
	}
	PResult = resp.PResult
	PLogonId = resp.PLogonId
	if uint32(resp.Status) != IcaApi.StatusSuccess {
		err = fmt.Errorf("RpcLogonIdFromWinStationName failed: %s", IcaApi.StatusString(uint32(resp.Status)))
	}
	return
}
