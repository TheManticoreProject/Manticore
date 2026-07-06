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

// apiOpenNodeExRequest carries the [in] parameters of ApiOpenNodeEx.
type apiOpenNodeExRequest struct {
	LpszNodeName    ndr.WSTR
	DwDesiredAccess ndr.DWORD
}

func (*apiOpenNodeExRequest) Opnum() uint16 { return clusapi.OpnumApiOpenNodeEx }

// apiOpenNodeExResponse carries the [out] parameters and return value of ApiOpenNodeEx.
type apiOpenNodeExResponse struct {
	LpdwGrantedAccess ndr.DWORD
	Status            ndr.DWORD
	Rpc_status        ndr.DWORD
	Handle            mscmrp.HNODE_RPC `ndr:"retval"`
}

// ApiOpenNodeEx calls ApiOpenNodeEx (opnum 118) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiOpenNodeEx(rpc ndr.Invoker, lpszNodeName ndr.WSTR, dwDesiredAccess ndr.DWORD) (Handle mscmrp.HNODE_RPC, LpdwGrantedAccess ndr.DWORD, Status ndr.DWORD, Rpc_status ndr.DWORD, err error) {
	req := &apiOpenNodeExRequest{
		LpszNodeName:    lpszNodeName,
		DwDesiredAccess: dwDesiredAccess,
	}
	var resp apiOpenNodeExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiOpenNodeEx: %w", err)
		return
	}
	Handle = resp.Handle
	LpdwGrantedAccess = resp.LpdwGrantedAccess
	Status = resp.Status
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiOpenNodeEx failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
