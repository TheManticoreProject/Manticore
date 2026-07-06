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

// apiGetNetInterfaceIdRequest carries the [in] parameters of ApiGetNetInterfaceId.
type apiGetNetInterfaceIdRequest struct {
	HNetInterface mscmrp.HNETINTERFACE_RPC
}

func (*apiGetNetInterfaceIdRequest) Opnum() uint16 { return clusapi.OpnumApiGetNetInterfaceId }

// apiGetNetInterfaceIdResponse carries the [out] parameters and return value of ApiGetNetInterfaceId.
type apiGetNetInterfaceIdResponse struct {
	PGuid      ndr.WSTR
	Rpc_status ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// ApiGetNetInterfaceId calls ApiGetNetInterfaceId (opnum 96) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiGetNetInterfaceId(rpc ndr.Invoker, hNetInterface mscmrp.HNETINTERFACE_RPC) (PGuid ndr.WSTR, Rpc_status ndr.DWORD, err error) {
	req := &apiGetNetInterfaceIdRequest{
		HNetInterface: hNetInterface,
	}
	var resp apiGetNetInterfaceIdResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiGetNetInterfaceId: %w", err)
		return
	}
	PGuid = resp.PGuid
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiGetNetInterfaceId failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
