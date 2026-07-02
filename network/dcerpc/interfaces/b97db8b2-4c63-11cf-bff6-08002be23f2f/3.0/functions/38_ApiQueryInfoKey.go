package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiQueryInfoKeyRequest carries the [in] parameters of ApiQueryInfoKey.
type apiQueryInfoKeyRequest struct {
	HKey mscmrp.HKEY_RPC
}

func (*apiQueryInfoKeyRequest) Opnum() uint16 { return clusapi.OpnumApiQueryInfoKey }

// apiQueryInfoKeyResponse carries the [out] parameters and return value of ApiQueryInfoKey.
type apiQueryInfoKeyResponse struct {
	LpcSubKeys             ndr.DWORD
	LpcbMaxSubKeyLen       ndr.DWORD
	LpcValues              ndr.DWORD
	LpcbMaxValueNameLen    ndr.DWORD
	LpcbMaxValueLen        ndr.DWORD
	LpcbSecurityDescriptor ndr.DWORD
	LpftLastWriteTime      dtyp.FILETIME
	Rpc_status             ndr.DWORD
	Status                 ndr.DWORD `ndr:"retval"`
}

// ApiQueryInfoKey calls ApiQueryInfoKey (opnum 38) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiQueryInfoKey(rpc ndr.Invoker, hKey mscmrp.HKEY_RPC) (LpcSubKeys ndr.DWORD, LpcbMaxSubKeyLen ndr.DWORD, LpcValues ndr.DWORD, LpcbMaxValueNameLen ndr.DWORD, LpcbMaxValueLen ndr.DWORD, LpcbSecurityDescriptor ndr.DWORD, LpftLastWriteTime dtyp.FILETIME, Rpc_status ndr.DWORD, err error) {
	req := &apiQueryInfoKeyRequest{
		HKey: hKey,
	}
	var resp apiQueryInfoKeyResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiQueryInfoKey: %w", err)
		return
	}
	LpcSubKeys = resp.LpcSubKeys
	LpcbMaxSubKeyLen = resp.LpcbMaxSubKeyLen
	LpcValues = resp.LpcValues
	LpcbMaxValueNameLen = resp.LpcbMaxValueNameLen
	LpcbMaxValueLen = resp.LpcbMaxValueLen
	LpcbSecurityDescriptor = resp.LpcbSecurityDescriptor
	LpftLastWriteTime = resp.LpftLastWriteTime
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiQueryInfoKey failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
