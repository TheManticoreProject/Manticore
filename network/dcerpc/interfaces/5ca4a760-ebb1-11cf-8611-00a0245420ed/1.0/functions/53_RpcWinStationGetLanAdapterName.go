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

// rpcWinStationGetLanAdapterNameRequest carries the [in] parameters of RpcWinStationGetLanAdapterName.
type rpcWinStationGetLanAdapterNameRequest struct {
	HServer    mststs.SERVER_HANDLE
	PdNameSize ndr.DWORD
	PPdName    []uint16 `ndr:"ref,size_is=PdNameSize"`
	LanAdapter ndr.DWORD
}

func (*rpcWinStationGetLanAdapterNameRequest) Opnum() uint16 {
	return IcaApi.OpnumRpcWinStationGetLanAdapterName
}

// rpcWinStationGetLanAdapterNameResponse carries the [out] parameters and return value of RpcWinStationGetLanAdapterName.
type rpcWinStationGetLanAdapterNameResponse struct {
	PResult      ndr.DWORD
	PLength      ndr.DWORD
	PpLanAdapter []uint16  `ndr:"unique,conformant"`
	Status       ndr.DWORD `ndr:"retval"`
}

// RpcWinStationGetLanAdapterName calls RpcWinStationGetLanAdapterName (opnum 53) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcWinStationGetLanAdapterName(rpc ndr.Invoker, hServer mststs.SERVER_HANDLE, pdNameSize ndr.DWORD, pPdName []uint16, lanAdapter ndr.DWORD) (PResult ndr.DWORD, PLength ndr.DWORD, PpLanAdapter []uint16, err error) {
	req := &rpcWinStationGetLanAdapterNameRequest{
		HServer:    hServer,
		PdNameSize: pdNameSize,
		PPdName:    pPdName,
		LanAdapter: lanAdapter,
	}
	var resp rpcWinStationGetLanAdapterNameResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcWinStationGetLanAdapterName: %w", err)
		return
	}
	PResult = resp.PResult
	PLength = resp.PLength
	PpLanAdapter = resp.PpLanAdapter
	if uint32(resp.Status) != IcaApi.StatusSuccess {
		err = fmt.Errorf("RpcWinStationGetLanAdapterName failed: %s", IcaApi.StatusString(uint32(resp.Status)))
	}
	return
}
