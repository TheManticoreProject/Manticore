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

// apiSetValueRequest carries the [in] parameters of ApiSetValue.
type apiSetValueRequest struct {
	HKey        mscmrp.HKEY_RPC
	LpValueName ndr.WSTR
	DwType      ndr.DWORD
	LpData      []uint8 `ndr:"ref,size_is=CbData"`
	CbData      ndr.DWORD
}

func (*apiSetValueRequest) Opnum() uint16 { return clusapi.OpnumApiSetValue }

// apiSetValueResponse carries the [out] parameters and return value of ApiSetValue.
type apiSetValueResponse struct {
	Rpc_status ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// ApiSetValue calls ApiSetValue (opnum 32) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiSetValue(rpc ndr.Invoker, hKey mscmrp.HKEY_RPC, lpValueName ndr.WSTR, dwType ndr.DWORD, lpData []uint8, cbData ndr.DWORD) (Rpc_status ndr.DWORD, err error) {
	req := &apiSetValueRequest{
		HKey:        hKey,
		LpValueName: lpValueName,
		DwType:      dwType,
		LpData:      lpData,
		CbData:      cbData,
	}
	var resp apiSetValueResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiSetValue: %w", err)
		return
	}
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiSetValue failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
