package functions

import (
	"fmt"

	rasrpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/20610036-fa22-11cf-9823-00a0c911e5df/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rasRpcGetInstalledProtocolsExRequest carries the [in] parameters of RasRpcGetInstalledProtocolsEx.
type rasRpcGetInstalledProtocolsExRequest struct {
	FRouter ndr.BOOL
	FRasCli ndr.BOOL
	FRasSrv ndr.BOOL
}

func (*rasRpcGetInstalledProtocolsExRequest) Opnum() uint16 {
	return rasrpc.OpnumRasRpcGetInstalledProtocolsEx
}

// rasRpcGetInstalledProtocolsExResponse carries the [out] parameters and return value of RasRpcGetInstalledProtocolsEx.
type rasRpcGetInstalledProtocolsExResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RasRpcGetInstalledProtocolsEx calls RasRpcGetInstalledProtocolsEx (opnum 14) ([MS-RRASM] — verify the parameter
// modeling and status handling).
func RasRpcGetInstalledProtocolsEx(rpc ndr.Invoker, fRouter ndr.BOOL, fRasCli ndr.BOOL, fRasSrv ndr.BOOL) (err error) {
	req := &rasRpcGetInstalledProtocolsExRequest{
		FRouter: fRouter,
		FRasCli: fRasCli,
		FRasSrv: fRasSrv,
	}
	var resp rasRpcGetInstalledProtocolsExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RasRpcGetInstalledProtocolsEx: %w", err)
		return
	}
	if uint32(resp.Status) != rasrpc.StatusSuccess {
		err = fmt.Errorf("RasRpcGetInstalledProtocolsEx failed: %s", rasrpc.StatusString(uint32(resp.Status)))
	}
	return
}
