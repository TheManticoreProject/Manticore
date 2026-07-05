package functions

import (
	"fmt"

	winreg "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/338cd001-2244-31f1-aaaa-900038001003/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rrp"
)

// baseRegLoadKeyRequest carries the [in] parameters of BaseRegLoadKey.
type baseRegLoadKeyRequest struct {
	HKey     msrrp.RPC_HKEY
	LpSubKey msrrp.RRP_UNICODE_STRING
	LpFile   msrrp.RRP_UNICODE_STRING
}

func (*baseRegLoadKeyRequest) Opnum() uint16 { return winreg.OpnumBaseRegLoadKey }

// baseRegLoadKeyResponse carries the [out] parameters and return value of BaseRegLoadKey.
type baseRegLoadKeyResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// BaseRegLoadKey calls BaseRegLoadKey (opnum 13) ([MS-RRP] — verify the parameter
// modeling and status handling).
func BaseRegLoadKey(rpc ndr.Invoker, hKey msrrp.RPC_HKEY, lpSubKey msrrp.RRP_UNICODE_STRING, lpFile msrrp.RRP_UNICODE_STRING) (err error) {
	req := &baseRegLoadKeyRequest{
		HKey:     hKey,
		LpSubKey: lpSubKey,
		LpFile:   lpFile,
	}
	var resp baseRegLoadKeyResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("BaseRegLoadKey: %w", err)
		return
	}
	if uint32(resp.Status) != winreg.StatusSuccess {
		err = fmt.Errorf("BaseRegLoadKey failed: %s", winreg.StatusString(uint32(resp.Status)))
	}
	return
}
