package functions

// IDL source: [MS-CMRP] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cmrp/e12b6d8f-c410-49d7-a27d-9992782a9027
// A fetched copy is kept at ms-cmrp.idl in the interface directory.

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiEnumValueRequest carries the [in] parameters of ApiEnumValue.
type apiEnumValueRequest struct {
	HKey     mscmrp.HKEY_RPC
	DwIndex  ndr.DWORD
	LpcbData ndr.DWORD
}

func (*apiEnumValueRequest) Opnum() uint16 { return clusapi.OpnumApiEnumValue }

// apiEnumValueResponse carries the [out] parameters and return value of ApiEnumValue.
type apiEnumValueResponse struct {
	LpValueName ndr.WSTR
	LpType      ndr.DWORD
	LpData      []uint8 `ndr:"ref,conformant"`
	LpcbData    ndr.DWORD
	TotalSize   ndr.DWORD
	Rpc_status  ndr.DWORD
	Status      ndr.DWORD `ndr:"retval"`
}

// ApiEnumValue calls ApiEnumValue (opnum 36) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiEnumValue(rpc ndr.Invoker, hKey mscmrp.HKEY_RPC, dwIndex ndr.DWORD, lpcbData ndr.DWORD) (LpValueName ndr.WSTR, LpType ndr.DWORD, LpData []uint8, LpcbData ndr.DWORD, TotalSize ndr.DWORD, Rpc_status ndr.DWORD, err error) {
	req := &apiEnumValueRequest{
		HKey:     hKey,
		DwIndex:  dwIndex,
		LpcbData: lpcbData,
	}
	var resp apiEnumValueResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiEnumValue: %w", err)
		return
	}
	LpValueName = resp.LpValueName
	LpType = resp.LpType
	LpData = resp.LpData
	LpcbData = resp.LpcbData
	TotalSize = resp.TotalSize
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiEnumValue failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
