package functions

import (
	"fmt"

	winreg "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/338cd001-2244-31f1-aaaa-900038001003/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/338cd001-2244-31f1-aaaa-900038001003/1.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// baseRegDeleteKeyExRequest carries the [in] parameters of BaseRegDeleteKeyEx.
type baseRegDeleteKeyExRequest struct {
	HKey       structures.RPC_HKEY
	LpSubKey   structures.RRP_UNICODE_STRING
	AccessMask ndr.DWORD
	Reserved   ndr.DWORD
}

func (*baseRegDeleteKeyExRequest) Opnum() uint16 { return winreg.OpnumBaseRegDeleteKeyEx }

// baseRegDeleteKeyExResponse carries the [out] parameters and return value of BaseRegDeleteKeyEx.
type baseRegDeleteKeyExResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// BaseRegDeleteKeyEx calls BaseRegDeleteKeyEx (opnum 35) ([MS-RRP] — verify the parameter
// modeling and status handling).
func BaseRegDeleteKeyEx(rpc ndr.Invoker, hKey structures.RPC_HKEY, lpSubKey structures.RRP_UNICODE_STRING, accessMask ndr.DWORD, reserved ndr.DWORD) (err error) {
	req := &baseRegDeleteKeyExRequest{
		HKey:       hKey,
		LpSubKey:   lpSubKey,
		AccessMask: accessMask,
		Reserved:   reserved,
	}
	var resp baseRegDeleteKeyExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("BaseRegDeleteKeyEx: %w", err)
		return
	}
	if uint32(resp.Status) != winreg.StatusSuccess {
		err = fmt.Errorf("BaseRegDeleteKeyEx failed: %s", winreg.StatusString(uint32(resp.Status)))
	}
	return
}
