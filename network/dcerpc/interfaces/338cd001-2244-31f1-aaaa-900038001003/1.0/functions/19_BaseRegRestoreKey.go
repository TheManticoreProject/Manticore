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

// baseRegRestoreKeyRequest carries the [in] parameters of BaseRegRestoreKey.
type baseRegRestoreKeyRequest struct {
	HKey   msrrp.RPC_HKEY
	LpFile msrrp.RRP_UNICODE_STRING
	Flags  ndr.DWORD
}

func (*baseRegRestoreKeyRequest) Opnum() uint16 { return winreg.OpnumBaseRegRestoreKey }

// baseRegRestoreKeyResponse carries the [out] parameters and return value of BaseRegRestoreKey.
type baseRegRestoreKeyResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// BaseRegRestoreKey calls BaseRegRestoreKey (opnum 19) ([MS-RRP] — verify the parameter
// modeling and status handling).
func BaseRegRestoreKey(rpc ndr.Invoker, hKey msrrp.RPC_HKEY, lpFile msrrp.RRP_UNICODE_STRING, flags ndr.DWORD) (err error) {
	req := &baseRegRestoreKeyRequest{
		HKey:   hKey,
		LpFile: lpFile,
		Flags:  flags,
	}
	var resp baseRegRestoreKeyResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("BaseRegRestoreKey: %w", err)
		return
	}
	if uint32(resp.Status) != winreg.StatusSuccess {
		err = fmt.Errorf("BaseRegRestoreKey failed: %s", winreg.StatusString(uint32(resp.Status)))
	}
	return
}
