package functions

import (
	"fmt"

	winreg "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/338cd001-2244-31f1-aaaa-900038001003/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rrp"
)

// baseRegSetValueRequest carries the [in] parameters of BaseRegSetValue.
type baseRegSetValueRequest struct {
	HKey        msrrp.RPC_HKEY
	LpValueName msrrp.RRP_UNICODE_STRING
	DwType      ndr.DWORD
	LpData      []uint8 `ndr:"ref,size_is=CbData"`
	CbData      ndr.DWORD
}

func (*baseRegSetValueRequest) Opnum() uint16 { return winreg.OpnumBaseRegSetValue }

// baseRegSetValueResponse carries the [out] parameters and return value of BaseRegSetValue.
type baseRegSetValueResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// BaseRegSetValue calls BaseRegSetValue (opnum 22) ([MS-RRP] — verify the parameter
// modeling and status handling).
func BaseRegSetValue(rpc ndr.Invoker, hKey msrrp.RPC_HKEY, lpValueName msrrp.RRP_UNICODE_STRING, dwType ndr.DWORD, lpData []uint8, cbData ndr.DWORD) (err error) {
	req := &baseRegSetValueRequest{
		HKey:        hKey,
		LpValueName: lpValueName,
		DwType:      dwType,
		LpData:      lpData,
		CbData:      cbData,
	}
	var resp baseRegSetValueResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("BaseRegSetValue: %w", err)
		return
	}
	if uint32(resp.Status) != winreg.StatusSuccess {
		err = fmt.Errorf("BaseRegSetValue failed: %s", winreg.StatusString(uint32(resp.Status)))
	}
	return
}
