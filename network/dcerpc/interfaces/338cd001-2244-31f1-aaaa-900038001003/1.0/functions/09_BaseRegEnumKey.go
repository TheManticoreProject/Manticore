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

// baseRegEnumKeyRequest carries the [in] parameters of BaseRegEnumKey.
type baseRegEnumKeyRequest struct {
	HKey              msrrp.RPC_HKEY
	DwIndex           ndr.DWORD
	LpNameIn          msrrp.RRP_UNICODE_STRING
	LpClassIn         *msrrp.RRP_UNICODE_STRING `ndr:"unique"`
	LpftLastWriteTime *msdtyp.FILETIME          `ndr:"unique"`
}

func (*baseRegEnumKeyRequest) Opnum() uint16 { return winreg.OpnumBaseRegEnumKey }

// baseRegEnumKeyResponse carries the [out] parameters and return value of BaseRegEnumKey.
type baseRegEnumKeyResponse struct {
	LpNameOut         msrrp.RRP_UNICODE_STRING
	LplpClassOut      *msdtyp.RPC_UNICODE_STRING `ndr:"unique"`
	LpftLastWriteTime *msdtyp.FILETIME           `ndr:"unique"`
	Status            ndr.DWORD                  `ndr:"retval"`
}

// BaseRegEnumKey calls BaseRegEnumKey (opnum 9) ([MS-RRP] — verify the parameter
// modeling and status handling).
func BaseRegEnumKey(rpc ndr.Invoker, hKey msrrp.RPC_HKEY, dwIndex ndr.DWORD, lpNameIn msrrp.RRP_UNICODE_STRING, lpClassIn *msrrp.RRP_UNICODE_STRING, lpftLastWriteTime *msdtyp.FILETIME) (LpNameOut msrrp.RRP_UNICODE_STRING, LplpClassOut *msdtyp.RPC_UNICODE_STRING, LpftLastWriteTime *msdtyp.FILETIME, err error) {
	req := &baseRegEnumKeyRequest{
		HKey:              hKey,
		DwIndex:           dwIndex,
		LpNameIn:          lpNameIn,
		LpClassIn:         lpClassIn,
		LpftLastWriteTime: lpftLastWriteTime,
	}
	var resp baseRegEnumKeyResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("BaseRegEnumKey: %w", err)
		return
	}
	LpNameOut = resp.LpNameOut
	LplpClassOut = resp.LplpClassOut
	LpftLastWriteTime = resp.LpftLastWriteTime
	if uint32(resp.Status) != winreg.StatusSuccess {
		err = fmt.Errorf("BaseRegEnumKey failed: %s", winreg.StatusString(uint32(resp.Status)))
	}
	return
}
