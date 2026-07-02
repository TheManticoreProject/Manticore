package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiGetRootKeyRequest carries the [in] parameters of ApiGetRootKey.
type apiGetRootKeyRequest struct {
	SamDesired ndr.DWORD
}

func (*apiGetRootKeyRequest) Opnum() uint16 { return clusapi.OpnumApiGetRootKey }

// apiGetRootKeyResponse carries the [out] parameters and return value of ApiGetRootKey.
type apiGetRootKeyResponse struct {
	Status     ndr.DWORD
	Rpc_status ndr.DWORD
	Handle     mscmrp.HKEY_RPC `ndr:"retval"`
}

// ApiGetRootKey calls ApiGetRootKey (opnum 28) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiGetRootKey(rpc ndr.Invoker, samDesired ndr.DWORD) (Handle mscmrp.HKEY_RPC, Status ndr.DWORD, Rpc_status ndr.DWORD, err error) {
	req := &apiGetRootKeyRequest{
		SamDesired: samDesired,
	}
	var resp apiGetRootKeyResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiGetRootKey: %w", err)
		return
	}
	Handle = resp.Handle
	Status = resp.Status
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiGetRootKey failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
