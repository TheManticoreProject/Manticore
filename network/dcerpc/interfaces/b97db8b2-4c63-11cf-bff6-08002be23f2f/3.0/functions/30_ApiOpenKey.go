package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiOpenKeyRequest carries the [in] parameters of ApiOpenKey.
type apiOpenKeyRequest struct {
	HKey       mscmrp.HKEY_RPC
	LpSubKey   ndr.WSTR
	SamDesired ndr.DWORD
}

func (*apiOpenKeyRequest) Opnum() uint16 { return clusapi.OpnumApiOpenKey }

// apiOpenKeyResponse carries the [out] parameters and return value of ApiOpenKey.
type apiOpenKeyResponse struct {
	Status     ndr.DWORD
	Rpc_status ndr.DWORD
	Handle     mscmrp.HKEY_RPC `ndr:"retval"`
}

// ApiOpenKey calls ApiOpenKey (opnum 30) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiOpenKey(rpc ndr.Invoker, hKey mscmrp.HKEY_RPC, lpSubKey ndr.WSTR, samDesired ndr.DWORD) (Handle mscmrp.HKEY_RPC, Status ndr.DWORD, Rpc_status ndr.DWORD, err error) {
	req := &apiOpenKeyRequest{
		HKey:       hKey,
		LpSubKey:   lpSubKey,
		SamDesired: samDesired,
	}
	var resp apiOpenKeyResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiOpenKey: %w", err)
		return
	}
	Handle = resp.Handle
	Status = resp.Status
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiOpenKey failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
