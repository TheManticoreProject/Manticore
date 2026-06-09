package functions

import (
	"fmt"

	winreg "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/338cd001-2244-31f1-aaaa-900038001003/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/338cd001-2244-31f1-aaaa-900038001003/1.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// baseRegGetVersionRequest carries the [in] parameters of BaseRegGetVersion.
type baseRegGetVersionRequest struct {
	HKey structures.RPC_HKEY
}

func (*baseRegGetVersionRequest) Opnum() uint16 { return winreg.OpnumBaseRegGetVersion }

// baseRegGetVersionResponse carries the [out] parameters and return value of BaseRegGetVersion.
type baseRegGetVersionResponse struct {
	LpdwVersion ndr.DWORD
	Status      ndr.DWORD `ndr:"retval"`
}

// BaseRegGetVersion calls BaseRegGetVersion (opnum 26) ([MS-RRP] — verify the parameter
// modeling and status handling).
func BaseRegGetVersion(rpc ndr.Invoker, hKey structures.RPC_HKEY) (LpdwVersion ndr.DWORD, err error) {
	req := &baseRegGetVersionRequest{
		HKey: hKey,
	}
	var resp baseRegGetVersionResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("BaseRegGetVersion: %w", err)
		return
	}
	LpdwVersion = resp.LpdwVersion
	if uint32(resp.Status) != winreg.StatusSuccess {
		err = fmt.Errorf("BaseRegGetVersion failed: %s", winreg.StatusString(uint32(resp.Status)))
	}
	return
}
