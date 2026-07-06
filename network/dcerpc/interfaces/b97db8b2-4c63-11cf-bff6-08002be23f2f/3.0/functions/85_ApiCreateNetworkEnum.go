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

// apiCreateNetworkEnumRequest carries the [in] parameters of ApiCreateNetworkEnum.
type apiCreateNetworkEnumRequest struct {
	HNetwork mscmrp.HNETWORK_RPC
	DwType   ndr.DWORD
}

func (*apiCreateNetworkEnumRequest) Opnum() uint16 { return clusapi.OpnumApiCreateNetworkEnum }

// apiCreateNetworkEnumResponse carries the [out] parameters and return value of ApiCreateNetworkEnum.
type apiCreateNetworkEnumResponse struct {
	ReturnEnum *mscmrp.ENUM_LIST `ndr:"unique"`
	Rpc_status ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// ApiCreateNetworkEnum calls ApiCreateNetworkEnum (opnum 85) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiCreateNetworkEnum(rpc ndr.Invoker, hNetwork mscmrp.HNETWORK_RPC, dwType ndr.DWORD) (ReturnEnum *mscmrp.ENUM_LIST, Rpc_status ndr.DWORD, err error) {
	req := &apiCreateNetworkEnumRequest{
		HNetwork: hNetwork,
		DwType:   dwType,
	}
	var resp apiCreateNetworkEnumResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiCreateNetworkEnum: %w", err)
		return
	}
	ReturnEnum = resp.ReturnEnum
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiCreateNetworkEnum failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
