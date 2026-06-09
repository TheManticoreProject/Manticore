package functions

import (
	"fmt"

	winreg "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/338cd001-2244-31f1-aaaa-900038001003/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/338cd001-2244-31f1-aaaa-900038001003/1.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// baseRegDeleteValueRequest carries the [in] parameters of BaseRegDeleteValue.
type baseRegDeleteValueRequest struct {
	HKey        structures.RPC_HKEY
	LpValueName structures.RRP_UNICODE_STRING
}

func (*baseRegDeleteValueRequest) Opnum() uint16 { return winreg.OpnumBaseRegDeleteValue }

// baseRegDeleteValueResponse carries the [out] parameters and return value of BaseRegDeleteValue.
type baseRegDeleteValueResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// BaseRegDeleteValue calls BaseRegDeleteValue (opnum 8) ([MS-RRP] — verify the parameter
// modeling and status handling).
func BaseRegDeleteValue(rpc ndr.Invoker, hKey structures.RPC_HKEY, lpValueName structures.RRP_UNICODE_STRING) (err error) {
	req := &baseRegDeleteValueRequest{
		HKey:        hKey,
		LpValueName: lpValueName,
	}
	var resp baseRegDeleteValueResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("BaseRegDeleteValue: %w", err)
		return
	}
	if uint32(resp.Status) != winreg.StatusSuccess {
		err = fmt.Errorf("BaseRegDeleteValue failed: %s", winreg.StatusString(uint32(resp.Status)))
	}
	return
}
