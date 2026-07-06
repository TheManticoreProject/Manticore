package functions

// IDL source: [MS-RRP] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rrp/47f3edf6-4c2d-45d8-ab5b-2dc077738903
// A fetched copy is kept at ms-rrp.idl in the interface directory.

import (
	"fmt"

	winreg "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/338cd001-2244-31f1-aaaa-900038001003/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
	msrrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rrp"
)

// baseRegEnumValueRequest carries the [in] parameters of BaseRegEnumValue.
type baseRegEnumValueRequest struct {
	HKey          msrrp.RPC_HKEY
	DwIndex       ndr.DWORD
	LpValueNameIn msrrp.RRP_UNICODE_STRING
	LpType        *ndr.DWORD `ndr:"unique"`
	LpData        []uint8    `ndr:"unique,varying"`
	LpcbData      *ndr.DWORD `ndr:"unique"`
	LpcbLen       *ndr.DWORD `ndr:"unique"`
}

func (*baseRegEnumValueRequest) Opnum() uint16 { return winreg.OpnumBaseRegEnumValue }

// baseRegEnumValueResponse carries the [out] parameters and return value of BaseRegEnumValue.
type baseRegEnumValueResponse struct {
	LpValueNameOut msdtyp.RPC_UNICODE_STRING
	LpType         *ndr.DWORD `ndr:"unique"`
	LpData         []uint8    `ndr:"unique,varying"`
	LpcbData       *ndr.DWORD `ndr:"unique"`
	LpcbLen        *ndr.DWORD `ndr:"unique"`
	Status         ndr.DWORD  `ndr:"retval"`
}

// BaseRegEnumValue calls BaseRegEnumValue (opnum 10) ([MS-RRP] — verify the parameter
// modeling and status handling).
func BaseRegEnumValue(rpc ndr.Invoker, hKey msrrp.RPC_HKEY, dwIndex ndr.DWORD, lpValueNameIn msrrp.RRP_UNICODE_STRING, lpType *ndr.DWORD, lpData []uint8, lpcbData *ndr.DWORD, lpcbLen *ndr.DWORD) (LpValueNameOut msdtyp.RPC_UNICODE_STRING, LpType *ndr.DWORD, LpData []uint8, LpcbData *ndr.DWORD, LpcbLen *ndr.DWORD, err error) {
	req := &baseRegEnumValueRequest{
		HKey:          hKey,
		DwIndex:       dwIndex,
		LpValueNameIn: lpValueNameIn,
		LpType:        lpType,
		LpData:        lpData,
		LpcbData:      lpcbData,
		LpcbLen:       lpcbLen,
	}
	var resp baseRegEnumValueResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("BaseRegEnumValue: %w", err)
		return
	}
	LpValueNameOut = resp.LpValueNameOut
	LpType = resp.LpType
	LpData = resp.LpData
	LpcbData = resp.LpcbData
	LpcbLen = resp.LpcbLen
	if uint32(resp.Status) != winreg.StatusSuccess {
		err = fmt.Errorf("BaseRegEnumValue failed: %s", winreg.StatusString(uint32(resp.Status)))
	}
	return
}
