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

// rpcWinStationShadowRequest carries the [in] parameters of RpcWinStationShadow.
type rpcWinStationShadowRequest struct {
	HServer           mststs.SERVER_HANDLE
	LogonId           ndr.DWORD
	PTargetServerName []uint16 `ndr:"ref,size_is=NameSize"`
	NameSize          ndr.DWORD
	TargetLogonId     ndr.DWORD
	HotKeyVk          uint8
	HotkeyModifiers   uint16
}

func (*rpcWinStationShadowRequest) Opnum() uint16 { return IcaApi.OpnumRpcWinStationShadow }

// rpcWinStationShadowResponse carries the [out] parameters and return value of RpcWinStationShadow.
type rpcWinStationShadowResponse struct {
	PResult ndr.DWORD
	Status  ndr.DWORD `ndr:"retval"`
}

// RpcWinStationShadow calls RpcWinStationShadow (opnum 17) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcWinStationShadow(rpc ndr.Invoker, hServer mststs.SERVER_HANDLE, logonId ndr.DWORD, pTargetServerName []uint16, nameSize ndr.DWORD, targetLogonId ndr.DWORD, hotKeyVk uint8, hotkeyModifiers uint16) (PResult ndr.DWORD, err error) {
	req := &rpcWinStationShadowRequest{
		HServer:           hServer,
		LogonId:           logonId,
		PTargetServerName: pTargetServerName,
		NameSize:          nameSize,
		TargetLogonId:     targetLogonId,
		HotKeyVk:          hotKeyVk,
		HotkeyModifiers:   hotkeyModifiers,
	}
	var resp rpcWinStationShadowResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcWinStationShadow: %w", err)
		return
	}
	PResult = resp.PResult
	if uint32(resp.Status) != IcaApi.StatusSuccess {
		err = fmt.Errorf("RpcWinStationShadow failed: %s", IcaApi.StatusString(uint32(resp.Status)))
	}
	return
}
