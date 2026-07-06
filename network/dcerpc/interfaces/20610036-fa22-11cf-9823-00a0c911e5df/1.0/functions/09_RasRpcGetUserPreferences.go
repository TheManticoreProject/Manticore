package functions

import (
	"fmt"

	rasrpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/20610036-fa22-11cf-9823-00a0c911e5df/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrrasm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rrasm"
)

// rasRpcGetUserPreferencesRequest carries the [in] parameters of RasRpcGetUserPreferences.
type rasRpcGetUserPreferencesRequest struct {
	PUser  msrrasm.RASRPC_PBUSER
	DwMode ndr.DWORD
}

func (*rasRpcGetUserPreferencesRequest) Opnum() uint16 { return rasrpc.OpnumRasRpcGetUserPreferences }

// rasRpcGetUserPreferencesResponse carries the [out] parameters and return value of RasRpcGetUserPreferences.
type rasRpcGetUserPreferencesResponse struct {
	PUser  msrrasm.RASRPC_PBUSER
	Status ndr.DWORD `ndr:"retval"`
}

// RasRpcGetUserPreferences calls RasRpcGetUserPreferences (opnum 9) ([MS-RRASM] — verify the parameter
// modeling and status handling).
func RasRpcGetUserPreferences(rpc ndr.Invoker, pUser msrrasm.RASRPC_PBUSER, dwMode ndr.DWORD) (PUser msrrasm.RASRPC_PBUSER, err error) {
	req := &rasRpcGetUserPreferencesRequest{
		PUser:  pUser,
		DwMode: dwMode,
	}
	var resp rasRpcGetUserPreferencesResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RasRpcGetUserPreferences: %w", err)
		return
	}
	PUser = resp.PUser
	if uint32(resp.Status) != rasrpc.StatusSuccess {
		err = fmt.Errorf("RasRpcGetUserPreferences failed: %s", rasrpc.StatusString(uint32(resp.Status)))
	}
	return
}
