package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
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
	LpftLastWriteTime dtyp.FILETIME
	Rpc_status        ndr.DWORD
	Status            ndr.DWORD `ndr:"retval"`
}

// ApiEnumKey calls ApiEnumKey (opnum 31) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiEnumKey(rpc ndr.Invoker, hKey mscmrp.HKEY_RPC, dwIndex ndr.DWORD) (KeyName ndr.WSTR, LpftLastWriteTime dtyp.FILETIME, Rpc_status ndr.DWORD, err error) {
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
