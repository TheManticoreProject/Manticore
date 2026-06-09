package functions

import (
	"fmt"

	winreg "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/338cd001-2244-31f1-aaaa-900038001003/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/338cd001-2244-31f1-aaaa-900038001003/1.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// baseRegReplaceKeyRequest carries the [in] parameters of BaseRegReplaceKey.
type baseRegReplaceKeyRequest struct {
	HKey      structures.RPC_HKEY
	LpSubKey  structures.RRP_UNICODE_STRING
	LpNewFile structures.RRP_UNICODE_STRING
	LpOldFile structures.RRP_UNICODE_STRING
}

func (*baseRegReplaceKeyRequest) Opnum() uint16 { return winreg.OpnumBaseRegReplaceKey }

// baseRegReplaceKeyResponse carries the [out] parameters and return value of BaseRegReplaceKey.
type baseRegReplaceKeyResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// BaseRegReplaceKey calls BaseRegReplaceKey (opnum 18) ([MS-RRP] — verify the parameter
// modeling and status handling).
func BaseRegReplaceKey(rpc ndr.Invoker, hKey structures.RPC_HKEY, lpSubKey structures.RRP_UNICODE_STRING, lpNewFile structures.RRP_UNICODE_STRING, lpOldFile structures.RRP_UNICODE_STRING) (err error) {
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
