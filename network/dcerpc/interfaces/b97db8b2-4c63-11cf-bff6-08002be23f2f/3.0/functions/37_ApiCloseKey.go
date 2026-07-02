package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiCloseKeyRequest carries the [in] parameters of ApiCloseKey.
type apiCloseKeyRequest struct {
	PKey mscmrp.HKEY_RPC
}

func (*apiCloseKeyRequest) Opnum() uint16 { return clusapi.OpnumApiCloseKey }

// apiCloseKeyResponse carries the [out] parameters and return value of ApiCloseKey.
type apiCloseKeyResponse struct {
	PKey   mscmrp.HKEY_RPC
	Status ndr.DWORD `ndr:"retval"`
}

// ApiCloseKey calls ApiCloseKey (opnum 37) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiCloseKey(rpc ndr.Invoker, pKey mscmrp.HKEY_RPC) (PKey mscmrp.HKEY_RPC, err error) {
	req := &apiCloseKeyRequest{
		PKey: pKey,
	}
	var resp apiCloseKeyResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiCloseKey: %w", err)
		return
	}
	PKey = resp.PKey
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiCloseKey failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
