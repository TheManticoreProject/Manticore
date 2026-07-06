package functions

// IDL source: [MS-CMRP] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cmrp/e12b6d8f-c410-49d7-a27d-9992782a9027
// A fetched copy is kept at ms-cmrp.idl in the interface directory.

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiEnumKeyRequest carries the [in] parameters of ApiEnumKey.
type apiEnumKeyRequest struct {
	HKey    mscmrp.HKEY_RPC
	DwIndex ndr.DWORD
}

func (*apiEnumKeyRequest) Opnum() uint16 { return clusapi.OpnumApiEnumKey }

// apiEnumKeyResponse carries the [out] parameters and return value of ApiEnumKey.
type apiEnumKeyResponse struct {
	KeyName           ndr.WSTR
	LpftLastWriteTime msdtyp.FILETIME
	Rpc_status        ndr.DWORD
	Status            ndr.DWORD `ndr:"retval"`
}

// ApiEnumKey calls ApiEnumKey (opnum 31) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiEnumKey(rpc ndr.Invoker, hKey mscmrp.HKEY_RPC, dwIndex ndr.DWORD) (KeyName ndr.WSTR, LpftLastWriteTime msdtyp.FILETIME, Rpc_status ndr.DWORD, err error) {
	req := &apiEnumKeyRequest{
		HKey:    hKey,
		DwIndex: dwIndex,
	}
	var resp apiEnumKeyResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiEnumKey: %w", err)
		return
	}
	KeyName = resp.KeyName
	LpftLastWriteTime = resp.LpftLastWriteTime
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiEnumKey failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
