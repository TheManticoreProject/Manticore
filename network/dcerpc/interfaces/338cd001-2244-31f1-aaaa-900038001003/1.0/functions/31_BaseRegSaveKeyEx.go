package functions

import (
	"fmt"

	winreg "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/338cd001-2244-31f1-aaaa-900038001003/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rrp"
)

// baseRegSaveKeyExRequest carries the [in] parameters of BaseRegSaveKeyEx.
type baseRegSaveKeyExRequest struct {
	HKey                msrrp.RPC_HKEY
	LpFile              msrrp.RRP_UNICODE_STRING
	PSecurityAttributes *msrrp.RPC_SECURITY_ATTRIBUTES `ndr:"unique"`
	Flags               ndr.DWORD
}

func (*baseRegSaveKeyExRequest) Opnum() uint16 { return winreg.OpnumBaseRegSaveKeyEx }

// baseRegSaveKeyExResponse carries the [out] parameters and return value of BaseRegSaveKeyEx.
type baseRegSaveKeyExResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// BaseRegSaveKeyEx calls BaseRegSaveKeyEx (opnum 31) ([MS-RRP] — verify the parameter
// modeling and status handling).
func BaseRegSaveKeyEx(rpc ndr.Invoker, hKey msrrp.RPC_HKEY, lpFile msrrp.RRP_UNICODE_STRING, pSecurityAttributes *msrrp.RPC_SECURITY_ATTRIBUTES, flags ndr.DWORD) (err error) {
	req := &baseRegSaveKeyExRequest{
		HKey:                hKey,
		LpFile:              lpFile,
		PSecurityAttributes: pSecurityAttributes,
		Flags:               flags,
	}
	var resp baseRegSaveKeyExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("BaseRegSaveKeyEx: %w", err)
		return
	}
	if uint32(resp.Status) != winreg.StatusSuccess {
		err = fmt.Errorf("BaseRegSaveKeyEx failed: %s", winreg.StatusString(uint32(resp.Status)))
	}
	return
}
