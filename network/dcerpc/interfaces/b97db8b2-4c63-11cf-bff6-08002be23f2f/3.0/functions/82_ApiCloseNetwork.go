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

// apiCloseNetworkRequest carries the [in] parameters of ApiCloseNetwork.
type apiCloseNetworkRequest struct {
	Network mscmrp.HNETWORK_RPC
}

func (*apiCloseNetworkRequest) Opnum() uint16 { return clusapi.OpnumApiCloseNetwork }

// apiCloseNetworkResponse carries the [out] parameters and return value of ApiCloseNetwork.
type apiCloseNetworkResponse struct {
	Network mscmrp.HNETWORK_RPC
	Status  ndr.DWORD `ndr:"retval"`
}

// ApiCloseNetwork calls ApiCloseNetwork (opnum 82) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiCloseNetwork(rpc ndr.Invoker, network mscmrp.HNETWORK_RPC) (Network mscmrp.HNETWORK_RPC, err error) {
	req := &apiCloseNetworkRequest{
		Network: network,
	}
	var resp apiCloseNetworkResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiCloseNetwork: %w", err)
		return
	}
	Network = resp.Network
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiCloseNetwork failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
