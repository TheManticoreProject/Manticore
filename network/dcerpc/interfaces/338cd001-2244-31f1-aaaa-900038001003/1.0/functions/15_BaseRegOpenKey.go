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

// baseRegOpenKeyRequest carries the [in] parameters of BaseRegOpenKey.
type baseRegOpenKeyRequest struct {
	HKey       msrrp.RPC_HKEY
	LpSubKey   msrrp.RRP_UNICODE_STRING
	DwOptions  ndr.DWORD
	SamDesired ndr.DWORD
}

func (*baseRegOpenKeyRequest) Opnum() uint16 { return winreg.OpnumBaseRegOpenKey }

// baseRegOpenKeyResponse carries the [out] parameters and return value of BaseRegOpenKey.
type baseRegOpenKeyResponse struct {
	PhkResult msrrp.PRPC_HKEY
	Status    ndr.DWORD `ndr:"retval"`
}

// BaseRegOpenKey calls BaseRegOpenKey (opnum 15) ([MS-RRP] — verify the parameter
// modeling and status handling).
func BaseRegOpenKey(rpc ndr.Invoker, hKey msrrp.RPC_HKEY, lpSubKey msrrp.RRP_UNICODE_STRING, dwOptions ndr.DWORD, samDesired ndr.DWORD) (PhkResult msrrp.PRPC_HKEY, err error) {
	req := &baseRegOpenKeyRequest{
		HKey:       hKey,
		LpSubKey:   lpSubKey,
		DwOptions:  dwOptions,
		SamDesired: samDesired,
	}
	var resp baseRegOpenKeyResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("BaseRegOpenKey: %w", err)
		return
	}
	PhkResult = resp.PhkResult
	if uint32(resp.Status) != winreg.StatusSuccess {
		err = fmt.Errorf("BaseRegOpenKey failed: %s", winreg.StatusString(uint32(resp.Status)))
	}
	return
}
