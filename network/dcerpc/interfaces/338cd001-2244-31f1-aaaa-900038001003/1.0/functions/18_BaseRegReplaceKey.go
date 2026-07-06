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

// baseRegReplaceKeyRequest carries the [in] parameters of BaseRegReplaceKey.
type baseRegReplaceKeyRequest struct {
	HKey      msrrp.RPC_HKEY
	LpSubKey  msrrp.RRP_UNICODE_STRING
	LpNewFile msrrp.RRP_UNICODE_STRING
	LpOldFile msrrp.RRP_UNICODE_STRING
}

func (*baseRegReplaceKeyRequest) Opnum() uint16 { return winreg.OpnumBaseRegReplaceKey }

// baseRegReplaceKeyResponse carries the [out] parameters and return value of BaseRegReplaceKey.
type baseRegReplaceKeyResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// BaseRegReplaceKey calls BaseRegReplaceKey (opnum 18) ([MS-RRP] — verify the parameter
// modeling and status handling).
func BaseRegReplaceKey(rpc ndr.Invoker, hKey msrrp.RPC_HKEY, lpSubKey msrrp.RRP_UNICODE_STRING, lpNewFile msrrp.RRP_UNICODE_STRING, lpOldFile msrrp.RRP_UNICODE_STRING) (err error) {
	req := &baseRegReplaceKeyRequest{
		HKey:      hKey,
		LpSubKey:  lpSubKey,
		LpNewFile: lpNewFile,
		LpOldFile: lpOldFile,
	}
	var resp baseRegReplaceKeyResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("BaseRegReplaceKey: %w", err)
		return
	}
	if uint32(resp.Status) != winreg.StatusSuccess {
		err = fmt.Errorf("BaseRegReplaceKey failed: %s", winreg.StatusString(uint32(resp.Status)))
	}
	return
}
