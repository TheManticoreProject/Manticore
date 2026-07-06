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

// baseRegDeleteKeyRequest carries the [in] parameters of BaseRegDeleteKey.
type baseRegDeleteKeyRequest struct {
	HKey     msrrp.RPC_HKEY
	LpSubKey msrrp.RRP_UNICODE_STRING
}

func (*baseRegDeleteKeyRequest) Opnum() uint16 { return winreg.OpnumBaseRegDeleteKey }

// baseRegDeleteKeyResponse carries the [out] parameters and return value of BaseRegDeleteKey.
type baseRegDeleteKeyResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// BaseRegDeleteKey calls BaseRegDeleteKey (opnum 7) ([MS-RRP] — verify the parameter
// modeling and status handling).
func BaseRegDeleteKey(rpc ndr.Invoker, hKey msrrp.RPC_HKEY, lpSubKey msrrp.RRP_UNICODE_STRING) (err error) {
	req := &baseRegDeleteKeyRequest{
		HKey:     hKey,
		LpSubKey: lpSubKey,
	}
	var resp baseRegDeleteKeyResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("BaseRegDeleteKey: %w", err)
		return
	}
	if uint32(resp.Status) != winreg.StatusSuccess {
		err = fmt.Errorf("BaseRegDeleteKey failed: %s", winreg.StatusString(uint32(resp.Status)))
	}
	return
}
