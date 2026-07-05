package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	winreg "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/338cd001-2244-31f1-aaaa-900038001003/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rrp"
)

// baseRegEnumKeyRequest carries the [in] parameters of BaseRegEnumKey.
type baseRegEnumKeyRequest struct {
	HKey              msrrp.RPC_HKEY
	DwIndex           ndr.DWORD
	LpNameIn          msrrp.RRP_UNICODE_STRING
	LpClassIn         *msrrp.RRP_UNICODE_STRING `ndr:"unique"`
	LpftLastWriteTime *dtyp.FILETIME            `ndr:"unique"`
}

func (*baseRegEnumKeyRequest) Opnum() uint16 { return winreg.OpnumBaseRegEnumKey }

// baseRegEnumKeyResponse carries the [out] parameters and return value of BaseRegEnumKey.
type baseRegEnumKeyResponse struct {
	LpNameOut         msrrp.RRP_UNICODE_STRING
	LplpClassOut      *dtyp.RPC_UNICODE_STRING `ndr:"unique"`
	LpftLastWriteTime *dtyp.FILETIME           `ndr:"unique"`
	Status            ndr.DWORD                `ndr:"retval"`
}

// BaseRegEnumKey calls BaseRegEnumKey (opnum 9) ([MS-RRP] — verify the parameter
// modeling and status handling).
func BaseRegEnumKey(rpc ndr.Invoker, hKey msrrp.RPC_HKEY, dwIndex ndr.DWORD, lpNameIn msrrp.RRP_UNICODE_STRING, lpClassIn *msrrp.RRP_UNICODE_STRING, lpftLastWriteTime *dtyp.FILETIME) (LpNameOut msrrp.RRP_UNICODE_STRING, LplpClassOut *dtyp.RPC_UNICODE_STRING, LpftLastWriteTime *dtyp.FILETIME, err error) {
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
