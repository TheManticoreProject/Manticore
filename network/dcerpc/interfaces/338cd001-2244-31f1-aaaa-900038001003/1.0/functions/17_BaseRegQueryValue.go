package functions

import (
	"fmt"

	winreg "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/338cd001-2244-31f1-aaaa-900038001003/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/338cd001-2244-31f1-aaaa-900038001003/1.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// baseRegQueryValueRequest carries the [in] parameters of BaseRegQueryValue.
type baseRegQueryValueRequest struct {
	HKey        structures.RPC_HKEY
	LpValueName structures.RRP_UNICODE_STRING
	LpType      *ndr.DWORD `ndr:"unique"`
	LpData      []uint8    `ndr:"unique,varying"`
	LpcbData    *ndr.DWORD `ndr:"unique"`
	LpcbLen     *ndr.DWORD `ndr:"unique"`
}

func (*baseRegQueryValueRequest) Opnum() uint16 { return winreg.OpnumBaseRegQueryValue }

// baseRegQueryValueResponse carries the [out] parameters and return value of BaseRegQueryValue.
type baseRegQueryValueResponse struct {
	LpType   *ndr.DWORD `ndr:"unique"`
	LpData   []uint8    `ndr:"unique,varying"`
	LpcbData *ndr.DWORD `ndr:"unique"`
	LpcbLen  *ndr.DWORD `ndr:"unique"`
	Status   ndr.DWORD  `ndr:"retval"`
}

// BaseRegQueryValue calls BaseRegQueryValue (opnum 17) ([MS-RRP] — verify the parameter
// modeling and status handling).
func BaseRegQueryValue(rpc ndr.Invoker, hKey structures.RPC_HKEY, lpValueName structures.RRP_UNICODE_STRING, lpType *ndr.DWORD, lpData []uint8, lpcbData *ndr.DWORD, lpcbLen *ndr.DWORD) (LpType *ndr.DWORD, LpData []uint8, LpcbData *ndr.DWORD, LpcbLen *ndr.DWORD, err error) {
	req := &baseRegQueryValueRequest{
		HKey:        hKey,
		LpValueName: lpValueName,
		LpType:      lpType,
		LpData:      lpData,
		LpcbData:    lpcbData,
		LpcbLen:     lpcbLen,
	}
	var resp baseRegQueryValueResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("BaseRegQueryValue: %w", err)
		return
	}
	LpType = resp.LpType
	LpData = resp.LpData
	LpcbData = resp.LpcbData
	LpcbLen = resp.LpcbLen
	if uint32(resp.Status) != winreg.StatusSuccess {
		err = fmt.Errorf("BaseRegQueryValue failed: %s", winreg.StatusString(uint32(resp.Status)))
	}
	return
}
