package functions

import (
	"fmt"

	rasrpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/20610036-fa22-11cf-9823-00a0c911e5df/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rasRpcGetVersionRequest carries the [in] parameters of RasRpcGetVersion.
type rasRpcGetVersionRequest struct {
	PdwVersion ndr.DWORD
}

func (*rasRpcGetVersionRequest) Opnum() uint16 { return rasrpc.OpnumRasRpcGetVersion }

// rasRpcGetVersionResponse carries the [out] parameters and return value of RasRpcGetVersion.
type rasRpcGetVersionResponse struct {
	PdwVersion ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// RasRpcGetVersion calls RasRpcGetVersion (opnum 15) ([MS-RRASM] — verify the parameter
// modeling and status handling).
func RasRpcGetVersion(rpc ndr.Invoker, pdwVersion ndr.DWORD) (PdwVersion ndr.DWORD, err error) {
	req := &rasRpcGetVersionRequest{
		PdwVersion: pdwVersion,
	}
	var resp rasRpcGetVersionResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RasRpcGetVersion: %w", err)
		return
	}
	PdwVersion = resp.PdwVersion
	if uint32(resp.Status) != rasrpc.StatusSuccess {
		err = fmt.Errorf("RasRpcGetVersion failed: %s", rasrpc.StatusString(uint32(resp.Status)))
	}
	return
}
