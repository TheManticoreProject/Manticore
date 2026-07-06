package functions

// IDL source: [MS-RRP] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rrp/47f3edf6-4c2d-45d8-ab5b-2dc077738903
// A fetched copy is kept at ms-rrp.idl in the interface directory.

import (
	"fmt"

	winreg "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/338cd001-2244-31f1-aaaa-900038001003/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msrrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rrp"
)

// baseRegQueryValueRequest carries the [in] parameters of BaseRegQueryValue.
//
// [MS-RRP] 3.1.5.17 declares lpData as
// [in, out, unique, size_is(*lpcbData), length_is(*lpcbLen)] LPBYTE: a unique pointer to a
// conformant-varying byte array whose maximum_count is *lpcbData (the buffer capacity the
// client offers) and whose actual_count is *lpcbLen (the number of valid input octets, 0 on
// a read). The two counts are independent of the Go slice length and of each other, so the
// tag names both sibling pointers explicitly — matching RPC_SECURITY_DESCRIPTOR's
// CbIn/CbOut modeling in this same interface. Without size_is/length_is the marshaller would
// derive both counts from len(LpData), transmitting an actual_count that contradicts *lpcbLen
// and a full input body on a value read, which a DC rejects with nca_s_fault_ndr.
type baseRegQueryValueRequest struct {
	HKey        msrrp.RPC_HKEY
	LpValueName msrrp.RRP_UNICODE_STRING
	LpType      *ndr.DWORD `ndr:"unique"`
	LpData      []uint8    `ndr:"unique,size_is=LpcbData,varying,length_is=LpcbLen"`
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
func BaseRegQueryValue(rpc ndr.Invoker, hKey msrrp.RPC_HKEY, lpValueName msrrp.RRP_UNICODE_STRING, lpType *ndr.DWORD, lpData []uint8, lpcbData *ndr.DWORD, lpcbLen *ndr.DWORD) (LpType *ndr.DWORD, LpData []uint8, LpcbData *ndr.DWORD, LpcbLen *ndr.DWORD, err error) {
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
