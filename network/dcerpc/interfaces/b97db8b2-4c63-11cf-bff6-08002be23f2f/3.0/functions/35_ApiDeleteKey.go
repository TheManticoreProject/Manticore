package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiDeleteKeyRequest carries the [in] parameters of ApiDeleteKey.
type apiDeleteKeyRequest struct {
	HKey     mscmrp.HKEY_RPC
	LpSubKey ndr.WSTR
}

func (*apiDeleteKeyRequest) Opnum() uint16 { return clusapi.OpnumApiDeleteKey }

// apiDeleteKeyResponse carries the [out] parameters and return value of ApiDeleteKey.
type apiDeleteKeyResponse struct {
	Rpc_status ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// ApiDeleteKey calls ApiDeleteKey (opnum 35) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiDeleteKey(rpc ndr.Invoker, hKey mscmrp.HKEY_RPC, lpSubKey ndr.WSTR) (Rpc_status ndr.DWORD, err error) {
	req := &apiDeleteKeyRequest{
		HKey:     hKey,
		LpSubKey: lpSubKey,
	}
	var resp apiDeleteKeyResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiDeleteKey: %w", err)
		return
	}
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiDeleteKey failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
