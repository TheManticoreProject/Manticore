package functions

// IDL source: [MS-CMRP] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cmrp/e12b6d8f-c410-49d7-a27d-9992782a9027
// A fetched copy is kept at ms-cmrp.idl in the interface directory.

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiCreateKeyRequest carries the [in] parameters of ApiCreateKey.
type apiCreateKeyRequest struct {
	HKey                 mscmrp.HKEY_RPC
	LpSubKey             ndr.WSTR
	DwOptions            ndr.DWORD
	SamDesired           ndr.DWORD
	LpSecurityAttributes *mscmrp.RPC_SECURITY_ATTRIBUTES `ndr:"unique"`
}

func (*apiCreateKeyRequest) Opnum() uint16 { return clusapi.OpnumApiCreateKey }

// apiCreateKeyResponse carries the [out] parameters and return value of ApiCreateKey.
type apiCreateKeyResponse struct {
	LpdwDisposition ndr.DWORD
	Status          ndr.DWORD
	Rpc_status      ndr.DWORD
	Handle          mscmrp.HKEY_RPC `ndr:"retval"`
}

// ApiCreateKey calls ApiCreateKey (opnum 29) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiCreateKey(rpc ndr.Invoker, hKey mscmrp.HKEY_RPC, lpSubKey ndr.WSTR, dwOptions ndr.DWORD, samDesired ndr.DWORD, lpSecurityAttributes *mscmrp.RPC_SECURITY_ATTRIBUTES) (Handle mscmrp.HKEY_RPC, LpdwDisposition ndr.DWORD, Status ndr.DWORD, Rpc_status ndr.DWORD, err error) {
	req := &apiCreateKeyRequest{
		HKey:                 hKey,
		LpSubKey:             lpSubKey,
		DwOptions:            dwOptions,
		SamDesired:           samDesired,
		LpSecurityAttributes: lpSecurityAttributes,
	}
	var resp apiCreateKeyResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiCreateKey: %w", err)
		return
	}
	Handle = resp.Handle
	LpdwDisposition = resp.LpdwDisposition
	Status = resp.Status
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiCreateKey failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
