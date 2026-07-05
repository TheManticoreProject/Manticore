package functions

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
