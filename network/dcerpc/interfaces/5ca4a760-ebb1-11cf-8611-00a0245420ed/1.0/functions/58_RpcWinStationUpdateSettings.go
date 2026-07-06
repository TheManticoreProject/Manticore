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

// rpcWinStationUpdateSettingsRequest carries the [in] parameters of RpcWinStationUpdateSettings.
type rpcWinStationUpdateSettingsRequest struct {
	HServer            mststs.SERVER_HANDLE
	SettingsClass      ndr.DWORD
	SettingsParameters ndr.DWORD
}

func (*rpcWinStationUpdateSettingsRequest) Opnum() uint16 {
	return IcaApi.OpnumRpcWinStationUpdateSettings
}

// rpcWinStationUpdateSettingsResponse carries the [out] parameters and return value of RpcWinStationUpdateSettings.
type rpcWinStationUpdateSettingsResponse struct {
	PResult ndr.DWORD
	Status  ndr.DWORD `ndr:"retval"`
}

// RpcWinStationUpdateSettings calls RpcWinStationUpdateSettings (opnum 58) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcWinStationUpdateSettings(rpc ndr.Invoker, hServer mststs.SERVER_HANDLE, settingsClass ndr.DWORD, settingsParameters ndr.DWORD) (PResult ndr.DWORD, err error) {
	req := &rpcWinStationUpdateSettingsRequest{
		HServer:            hServer,
		SettingsClass:      settingsClass,
		SettingsParameters: settingsParameters,
	}
	var resp rpcWinStationUpdateSettingsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcWinStationUpdateSettings: %w", err)
		return
	}
	PResult = resp.PResult
	if uint32(resp.Status) != IcaApi.StatusSuccess {
		err = fmt.Errorf("RpcWinStationUpdateSettings failed: %s", IcaApi.StatusString(uint32(resp.Status)))
	}
	return
}
