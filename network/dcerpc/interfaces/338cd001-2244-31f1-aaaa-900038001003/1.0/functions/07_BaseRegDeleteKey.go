package functions

import (
	"fmt"

	winreg "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/338cd001-2244-31f1-aaaa-900038001003/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/338cd001-2244-31f1-aaaa-900038001003/1.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// baseRegDeleteKeyRequest carries the [in] parameters of BaseRegDeleteKey.
type baseRegDeleteKeyRequest struct {
	HKey     structures.RPC_HKEY
	LpSubKey structures.RRP_UNICODE_STRING
}

func (*baseRegDeleteKeyRequest) Opnum() uint16 { return winreg.OpnumBaseRegDeleteKey }

// baseRegDeleteKeyResponse carries the [out] parameters and return value of BaseRegDeleteKey.
type baseRegDeleteKeyResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// BaseRegDeleteKey calls BaseRegDeleteKey (opnum 7) ([MS-RRP] — verify the parameter
// modeling and status handling).
func BaseRegDeleteKey(rpc ndr.Invoker, hKey structures.RPC_HKEY, lpSubKey structures.RRP_UNICODE_STRING) (err error) {
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
