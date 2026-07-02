package functions

import (
	"fmt"

	RemoteFW "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6b5bdd1e-528c-422c-af8c-a4079be4fe48/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msfasp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-fasp"
)

// rRPC_FWEnumNetworksRequest carries the [in] parameters of RRPC_FWEnumNetworks.
type rRPC_FWEnumNetworksRequest struct {
	HPolicyStore msfasp.FW_POLICY_STORE_HANDLE
}

func (*rRPC_FWEnumNetworksRequest) Opnum() uint16 { return RemoteFW.OpnumRRPC_FWEnumNetworks }

// rRPC_FWEnumNetworksResponse carries the [out] parameters and return value of RRPC_FWEnumNetworks.
type rRPC_FWEnumNetworksResponse struct {
	PdwNumNetworks ndr.DWORD
	PpNetworks     []*msfasp.FW_NETWORK `ndr:"elem=unique,ref,conformant"`
	Status         ndr.DWORD            `ndr:"retval"`
}

// RRPC_FWEnumNetworks calls RRPC_FWEnumNetworks (opnum 42) ([MS-FASP] — verify the parameter
// modeling and status handling).
func RRPC_FWEnumNetworks(rpc ndr.Invoker, hPolicyStore msfasp.FW_POLICY_STORE_HANDLE) (PdwNumNetworks ndr.DWORD, PpNetworks []*msfasp.FW_NETWORK, err error) {
	req := &rRPC_FWEnumNetworksRequest{
		HPolicyStore: hPolicyStore,
	}
	var resp rRPC_FWEnumNetworksResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRPC_FWEnumNetworks: %w", err)
		return
	}
	PdwNumNetworks = resp.PdwNumNetworks
	PpNetworks = resp.PpNetworks
	if uint32(resp.Status) != RemoteFW.StatusSuccess {
		err = fmt.Errorf("RRPC_FWEnumNetworks failed: %s", RemoteFW.StatusString(uint32(resp.Status)))
	}
	return
}
