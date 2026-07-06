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

// apiOpenGroupSetRequest carries the [in] parameters of ApiOpenGroupSet.
type apiOpenGroupSetRequest struct {
	LpszGroupSetName ndr.WSTR
}

func (*apiOpenGroupSetRequest) Opnum() uint16 { return clusapi.OpnumApiOpenGroupSet }

// apiOpenGroupSetResponse carries the [out] parameters and return value of ApiOpenGroupSet.
type apiOpenGroupSetResponse struct {
	Status     ndr.DWORD
	Rpc_status ndr.DWORD
	Handle     mscmrp.HGROUPSET_RPC `ndr:"retval"`
}

// ApiOpenGroupSet calls ApiOpenGroupSet (opnum 164) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiOpenGroupSet(rpc ndr.Invoker, lpszGroupSetName ndr.WSTR) (Handle mscmrp.HGROUPSET_RPC, Status ndr.DWORD, Rpc_status ndr.DWORD, err error) {
	req := &apiOpenGroupSetRequest{
		LpszGroupSetName: lpszGroupSetName,
	}
	var resp apiOpenGroupSetResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiOpenGroupSet: %w", err)
		return
	}
	Handle = resp.Handle
	Status = resp.Status
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiOpenGroupSet failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
