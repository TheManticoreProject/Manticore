package functions

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiOpenClusterExRequest carries the [in] parameters of ApiOpenClusterEx.
type apiOpenClusterExRequest struct {
	DwDesiredAccess ndr.DWORD
}

func (*apiOpenClusterExRequest) Opnum() uint16 { return clusapi.OpnumApiOpenClusterEx }

// apiOpenClusterExResponse carries the [out] parameters and return value of ApiOpenClusterEx.
type apiOpenClusterExResponse struct {
	LpdwGrantedAccess ndr.DWORD
	Status            ndr.DWORD
	Handle            mscmrp.HCLUSTER_RPC `ndr:"retval"`
}

// ApiOpenClusterEx calls ApiOpenClusterEx (opnum 117) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiOpenClusterEx(rpc ndr.Invoker, dwDesiredAccess ndr.DWORD) (Handle mscmrp.HCLUSTER_RPC, LpdwGrantedAccess ndr.DWORD, Status ndr.DWORD, err error) {
	req := &apiOpenClusterExRequest{
		DwDesiredAccess: dwDesiredAccess,
	}
	var resp apiOpenClusterExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiOpenClusterEx: %w", err)
		return
	}
	Handle = resp.Handle
	LpdwGrantedAccess = resp.LpdwGrantedAccess
	Status = resp.Status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiOpenClusterEx failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
