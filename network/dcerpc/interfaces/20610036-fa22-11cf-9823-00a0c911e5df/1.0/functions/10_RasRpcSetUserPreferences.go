package functions

// IDL source: [MS-RRASM] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rrasm/8e6e89fb-9c80-4a9a-a222-d7d8948244bb
// A fetched copy is kept at ms-rrasm.idl in the interface directory.

import (
	"fmt"

	rasrpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/20610036-fa22-11cf-9823-00a0c911e5df/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrrasm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rrasm"
)

// rasRpcSetUserPreferencesRequest carries the [in] parameters of RasRpcSetUserPreferences.
type rasRpcSetUserPreferencesRequest struct {
	PUser  msrrasm.RASRPC_PBUSER
	DwMode ndr.DWORD
}

func (*rasRpcSetUserPreferencesRequest) Opnum() uint16 { return rasrpc.OpnumRasRpcSetUserPreferences }

// rasRpcSetUserPreferencesResponse carries the [out] parameters and return value of RasRpcSetUserPreferences.
type rasRpcSetUserPreferencesResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RasRpcSetUserPreferences calls RasRpcSetUserPreferences (opnum 10) ([MS-RRASM] — verify the parameter
// modeling and status handling).
func RasRpcSetUserPreferences(rpc ndr.Invoker, pUser msrrasm.RASRPC_PBUSER, dwMode ndr.DWORD) (err error) {
	req := &rasRpcSetUserPreferencesRequest{
		PUser:  pUser,
		DwMode: dwMode,
	}
	var resp rasRpcSetUserPreferencesResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RasRpcSetUserPreferences: %w", err)
		return
	}
	if uint32(resp.Status) != rasrpc.StatusSuccess {
		err = fmt.Errorf("RasRpcSetUserPreferences failed: %s", rasrpc.StatusString(uint32(resp.Status)))
	}
	return
}
