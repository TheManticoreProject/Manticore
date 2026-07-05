package functions

import (
	"fmt"

	winreg "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/338cd001-2244-31f1-aaaa-900038001003/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rrp"
)

// baseRegUnLoadKeyRequest carries the [in] parameters of BaseRegUnLoadKey.
type baseRegUnLoadKeyRequest struct {
	HKey     msrrp.RPC_HKEY
	LpSubKey msrrp.RRP_UNICODE_STRING
}

func (*baseRegUnLoadKeyRequest) Opnum() uint16 { return winreg.OpnumBaseRegUnLoadKey }

// baseRegUnLoadKeyResponse carries the [out] parameters and return value of BaseRegUnLoadKey.
type baseRegUnLoadKeyResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// BaseRegUnLoadKey calls BaseRegUnLoadKey (opnum 23) ([MS-RRP] — verify the parameter
// modeling and status handling).
func BaseRegUnLoadKey(rpc ndr.Invoker, hKey msrrp.RPC_HKEY, lpSubKey msrrp.RRP_UNICODE_STRING) (err error) {
	req := &baseRegUnLoadKeyRequest{
		HKey:     hKey,
		LpSubKey: lpSubKey,
	}
	var resp baseRegUnLoadKeyResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("BaseRegUnLoadKey: %w", err)
		return
	}
	if uint32(resp.Status) != winreg.StatusSuccess {
		err = fmt.Errorf("BaseRegUnLoadKey failed: %s", winreg.StatusString(uint32(resp.Status)))
	}
	return
}
