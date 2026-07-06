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

// apiAddNotifyKeyRequest carries the [in] parameters of ApiAddNotifyKey.
type apiAddNotifyKeyRequest struct {
	HNotify      mscmrp.HNOTIFY_RPC
	HKey         mscmrp.HKEY_RPC
	DwNotifyKey  ndr.DWORD
	Filter       ndr.DWORD
	WatchSubTree ndr.BOOL
}

func (*apiAddNotifyKeyRequest) Opnum() uint16 { return clusapi.OpnumApiAddNotifyKey }

// apiAddNotifyKeyResponse carries the [out] parameters and return value of ApiAddNotifyKey.
type apiAddNotifyKeyResponse struct {
	Rpc_status ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// ApiAddNotifyKey calls ApiAddNotifyKey (opnum 61) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiAddNotifyKey(rpc ndr.Invoker, hNotify mscmrp.HNOTIFY_RPC, hKey mscmrp.HKEY_RPC, dwNotifyKey ndr.DWORD, filter ndr.DWORD, watchSubTree ndr.BOOL) (Rpc_status ndr.DWORD, err error) {
	req := &apiAddNotifyKeyRequest{
		HNotify:      hNotify,
		HKey:         hKey,
		DwNotifyKey:  dwNotifyKey,
		Filter:       filter,
		WatchSubTree: watchSubTree,
	}
	var resp apiAddNotifyKeyResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiAddNotifyKey: %w", err)
		return
	}
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiAddNotifyKey failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
