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

// apiQueryValueRequest carries the [in] parameters of ApiQueryValue.
type apiQueryValueRequest struct {
	HKey        mscmrp.HKEY_RPC
	LpValueName ndr.WSTR
	CbData      ndr.DWORD
}

func (*apiQueryValueRequest) Opnum() uint16 { return clusapi.OpnumApiQueryValue }

// apiQueryValueResponse carries the [out] parameters and return value of ApiQueryValue.
type apiQueryValueResponse struct {
	LpValueType  ndr.DWORD
	LpData       []uint8 `ndr:"ref,size_is=CbData"`
	LpcbRequired ndr.DWORD
	Rpc_status   ndr.DWORD
	Status       ndr.DWORD `ndr:"retval"`
}

// ApiQueryValue calls ApiQueryValue (opnum 34) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiQueryValue(rpc ndr.Invoker, hKey mscmrp.HKEY_RPC, lpValueName ndr.WSTR, cbData ndr.DWORD) (LpValueType ndr.DWORD, LpData []uint8, LpcbRequired ndr.DWORD, Rpc_status ndr.DWORD, err error) {
	req := &apiQueryValueRequest{
		HKey:        hKey,
		LpValueName: lpValueName,
		CbData:      cbData,
	}
	var resp apiQueryValueResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiQueryValue: %w", err)
		return
	}
	LpValueType = resp.LpValueType
	LpData = resp.LpData
	LpcbRequired = resp.LpcbRequired
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiQueryValue failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
