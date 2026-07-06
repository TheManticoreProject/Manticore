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

// apiOpenNetInterfaceExRequest carries the [in] parameters of ApiOpenNetInterfaceEx.
type apiOpenNetInterfaceExRequest struct {
	LpszNetInterfaceName ndr.WSTR
	DwDesiredAccess      ndr.DWORD
}

func (*apiOpenNetInterfaceExRequest) Opnum() uint16 { return clusapi.OpnumApiOpenNetInterfaceEx }

// apiOpenNetInterfaceExResponse carries the [out] parameters and return value of ApiOpenNetInterfaceEx.
type apiOpenNetInterfaceExResponse struct {
	LpdwGrantedAccess ndr.DWORD
	Status            ndr.DWORD
	Rpc_status        ndr.DWORD
	Handle            mscmrp.HNETINTERFACE_RPC `ndr:"retval"`
}

// ApiOpenNetInterfaceEx calls ApiOpenNetInterfaceEx (opnum 122) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiOpenNetInterfaceEx(rpc ndr.Invoker, lpszNetInterfaceName ndr.WSTR, dwDesiredAccess ndr.DWORD) (Handle mscmrp.HNETINTERFACE_RPC, LpdwGrantedAccess ndr.DWORD, Status ndr.DWORD, Rpc_status ndr.DWORD, err error) {
	req := &apiOpenNetInterfaceExRequest{
		LpszNetInterfaceName: lpszNetInterfaceName,
		DwDesiredAccess:      dwDesiredAccess,
	}
	var resp apiOpenNetInterfaceExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiOpenNetInterfaceEx: %w", err)
		return
	}
	Handle = resp.Handle
	LpdwGrantedAccess = resp.LpdwGrantedAccess
	Status = resp.Status
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiOpenNetInterfaceEx failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
