package functions

// IDL source: [MS-RRP] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rrp/47f3edf6-4c2d-45d8-ab5b-2dc077738903
// A fetched copy is kept at ms-rrp.idl in the interface directory.

import (
	"fmt"

	winreg "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/338cd001-2244-31f1-aaaa-900038001003/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
	msrrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rrp"
)

// baseRegQueryInfoKeyRequest carries the [in] parameters of BaseRegQueryInfoKey.
type baseRegQueryInfoKeyRequest struct {
	HKey      msrrp.RPC_HKEY
	LpClassIn msrrp.RRP_UNICODE_STRING
}

func (*baseRegQueryInfoKeyRequest) Opnum() uint16 { return winreg.OpnumBaseRegQueryInfoKey }

// baseRegQueryInfoKeyResponse carries the [out] parameters and return value of BaseRegQueryInfoKey.
type baseRegQueryInfoKeyResponse struct {
	LpClassOut             msdtyp.RPC_UNICODE_STRING
	LpcSubKeys             ndr.DWORD
	LpcbMaxSubKeyLen       ndr.DWORD
	LpcbMaxClassLen        ndr.DWORD
	LpcValues              ndr.DWORD
	LpcbMaxValueNameLen    ndr.DWORD
	LpcbMaxValueLen        ndr.DWORD
	LpcbSecurityDescriptor ndr.DWORD
	LpftLastWriteTime      msdtyp.FILETIME
	Status                 ndr.DWORD `ndr:"retval"`
}

// BaseRegQueryInfoKey calls BaseRegQueryInfoKey (opnum 16) ([MS-RRP] — verify the parameter
// modeling and status handling).
func BaseRegQueryInfoKey(rpc ndr.Invoker, hKey msrrp.RPC_HKEY, lpClassIn msrrp.RRP_UNICODE_STRING) (LpClassOut msdtyp.RPC_UNICODE_STRING, LpcSubKeys ndr.DWORD, LpcbMaxSubKeyLen ndr.DWORD, LpcbMaxClassLen ndr.DWORD, LpcValues ndr.DWORD, LpcbMaxValueNameLen ndr.DWORD, LpcbMaxValueLen ndr.DWORD, LpcbSecurityDescriptor ndr.DWORD, LpftLastWriteTime msdtyp.FILETIME, err error) {
	req := &baseRegQueryInfoKeyRequest{
		HKey:      hKey,
		LpClassIn: lpClassIn,
	}
	var resp baseRegQueryInfoKeyResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("BaseRegQueryInfoKey: %w", err)
		return
	}
	LpClassOut = resp.LpClassOut
	LpcSubKeys = resp.LpcSubKeys
	LpcbMaxSubKeyLen = resp.LpcbMaxSubKeyLen
	LpcbMaxClassLen = resp.LpcbMaxClassLen
	LpcValues = resp.LpcValues
	LpcbMaxValueNameLen = resp.LpcbMaxValueNameLen
	LpcbMaxValueLen = resp.LpcbMaxValueLen
	LpcbSecurityDescriptor = resp.LpcbSecurityDescriptor
	LpftLastWriteTime = resp.LpftLastWriteTime
	if uint32(resp.Status) != winreg.StatusSuccess {
		err = fmt.Errorf("BaseRegQueryInfoKey failed: %s", winreg.StatusString(uint32(resp.Status)))
	}
	return
}
