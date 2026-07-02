package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiDeleteValueRequest carries the [in] parameters of ApiDeleteValue.
type apiDeleteValueRequest struct {
	HKey        mscmrp.HKEY_RPC
	LpValueName ndr.WSTR
}

func (*apiDeleteValueRequest) Opnum() uint16 { return clusapi.OpnumApiDeleteValue }

// apiDeleteValueResponse carries the [out] parameters and return value of ApiDeleteValue.
type apiDeleteValueResponse struct {
	Rpc_status ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// ApiDeleteValue calls ApiDeleteValue (opnum 33) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiDeleteValue(rpc ndr.Invoker, hKey mscmrp.HKEY_RPC, lpValueName ndr.WSTR) (Rpc_status ndr.DWORD, err error) {
	req := &apiDeleteValueRequest{
		HKey:        hKey,
		LpValueName: lpValueName,
	}
	var resp apiDeleteValueResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiDeleteValue: %w", err)
		return
	}
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiDeleteValue failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
