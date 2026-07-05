package functions

import (
	"fmt"

	winreg "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/338cd001-2244-31f1-aaaa-900038001003/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rrp"
)

// baseRegSaveKeyRequest carries the [in] parameters of BaseRegSaveKey.
type baseRegSaveKeyRequest struct {
	HKey                msrrp.RPC_HKEY
	LpFile              msrrp.RRP_UNICODE_STRING
	PSecurityAttributes *msrrp.RPC_SECURITY_ATTRIBUTES `ndr:"unique"`
}

func (*baseRegSaveKeyRequest) Opnum() uint16 { return winreg.OpnumBaseRegSaveKey }

// baseRegSaveKeyResponse carries the [out] parameters and return value of BaseRegSaveKey.
type baseRegSaveKeyResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// BaseRegSaveKey calls BaseRegSaveKey (opnum 20) ([MS-RRP] — verify the parameter
// modeling and status handling).
func BaseRegSaveKey(rpc ndr.Invoker, hKey msrrp.RPC_HKEY, lpFile msrrp.RRP_UNICODE_STRING, pSecurityAttributes *msrrp.RPC_SECURITY_ATTRIBUTES) (err error) {
	req := &baseRegSaveKeyRequest{
		HKey:                hKey,
		LpFile:              lpFile,
		PSecurityAttributes: pSecurityAttributes,
	}
	var resp baseRegSaveKeyResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("BaseRegSaveKey: %w", err)
		return
	}
	if uint32(resp.Status) != winreg.StatusSuccess {
		err = fmt.Errorf("BaseRegSaveKey failed: %s", winreg.StatusString(uint32(resp.Status)))
	}
	return
}
