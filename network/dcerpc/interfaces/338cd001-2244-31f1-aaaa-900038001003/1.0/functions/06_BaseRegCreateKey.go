package functions

// IDL source: [MS-RRP] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rrp/47f3edf6-4c2d-45d8-ab5b-2dc077738903
// A fetched copy is kept at ms-rrp.idl in the interface directory.

import (
	"fmt"

	winreg "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/338cd001-2244-31f1-aaaa-900038001003/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rrp"
)

// baseRegCreateKeyRequest carries the [in] parameters of BaseRegCreateKey.
type baseRegCreateKeyRequest struct {
	HKey                 msrrp.RPC_HKEY
	LpSubKey             msrrp.RRP_UNICODE_STRING
	LpClass              msrrp.RRP_UNICODE_STRING
	DwOptions            ndr.DWORD
	SamDesired           ndr.DWORD
	LpSecurityAttributes *msrrp.RPC_SECURITY_ATTRIBUTES `ndr:"unique"`
	LpdwDisposition      *ndr.DWORD                     `ndr:"unique"`
}

func (*baseRegCreateKeyRequest) Opnum() uint16 { return winreg.OpnumBaseRegCreateKey }

// baseRegCreateKeyResponse carries the [out] parameters and return value of BaseRegCreateKey.
type baseRegCreateKeyResponse struct {
	PhkResult       msrrp.PRPC_HKEY
	LpdwDisposition *ndr.DWORD `ndr:"unique"`
	Status          ndr.DWORD  `ndr:"retval"`
}

// BaseRegCreateKey calls BaseRegCreateKey (opnum 6) ([MS-RRP] — verify the parameter
// modeling and status handling).
func BaseRegCreateKey(rpc ndr.Invoker, hKey msrrp.RPC_HKEY, lpSubKey msrrp.RRP_UNICODE_STRING, lpClass msrrp.RRP_UNICODE_STRING, dwOptions ndr.DWORD, samDesired ndr.DWORD, lpSecurityAttributes *msrrp.RPC_SECURITY_ATTRIBUTES, lpdwDisposition *ndr.DWORD) (PhkResult msrrp.PRPC_HKEY, LpdwDisposition *ndr.DWORD, err error) {
	req := &baseRegCreateKeyRequest{
		HKey:                 hKey,
		LpSubKey:             lpSubKey,
		LpClass:              lpClass,
		DwOptions:            dwOptions,
		SamDesired:           samDesired,
		LpSecurityAttributes: lpSecurityAttributes,
		LpdwDisposition:      lpdwDisposition,
	}
	var resp baseRegCreateKeyResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("BaseRegCreateKey: %w", err)
		return
	}
	PhkResult = resp.PhkResult
	LpdwDisposition = resp.LpdwDisposition
	if uint32(resp.Status) != winreg.StatusSuccess {
		err = fmt.Errorf("BaseRegCreateKey failed: %s", winreg.StatusString(uint32(resp.Status)))
	}
	return
}
