package functions

import (
	"fmt"

	winreg "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/338cd001-2244-31f1-aaaa-900038001003/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rrp"
)

// baseRegCloseKeyRequest carries the [in] parameters of BaseRegCloseKey.
type baseRegCloseKeyRequest struct {
	HKey msrrp.PRPC_HKEY
}

func (*baseRegCloseKeyRequest) Opnum() uint16 { return winreg.OpnumBaseRegCloseKey }

// baseRegCloseKeyResponse carries the [out] parameters and return value of BaseRegCloseKey.
type baseRegCloseKeyResponse struct {
	HKey   msrrp.PRPC_HKEY
	Status ndr.DWORD `ndr:"retval"`
}

// BaseRegCloseKey calls BaseRegCloseKey (opnum 5) ([MS-RRP] — verify the parameter
// modeling and status handling).
func BaseRegCloseKey(rpc ndr.Invoker, hKey msrrp.PRPC_HKEY) (HKey msrrp.PRPC_HKEY, err error) {
	req := &baseRegCloseKeyRequest{
		HKey: hKey,
	}
	var resp baseRegCloseKeyResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("BaseRegCloseKey: %w", err)
		return
	}
	HKey = resp.HKey
	if uint32(resp.Status) != winreg.StatusSuccess {
		err = fmt.Errorf("BaseRegCloseKey failed: %s", winreg.StatusString(uint32(resp.Status)))
	}
	return
}
