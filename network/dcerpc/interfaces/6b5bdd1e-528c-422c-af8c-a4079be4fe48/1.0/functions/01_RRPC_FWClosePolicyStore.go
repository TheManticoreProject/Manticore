package functions

import (
	"fmt"

	RemoteFW "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6b5bdd1e-528c-422c-af8c-a4079be4fe48/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msfasp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-fasp"
)

// rRPC_FWClosePolicyStoreRequest carries the [in] parameters of RRPC_FWClosePolicyStore.
type rRPC_FWClosePolicyStoreRequest struct {
	PhPolicyStore msfasp.PFW_POLICY_STORE_HANDLE
}

func (*rRPC_FWClosePolicyStoreRequest) Opnum() uint16 { return RemoteFW.OpnumRRPC_FWClosePolicyStore }

// rRPC_FWClosePolicyStoreResponse carries the [out] parameters and return value of RRPC_FWClosePolicyStore.
type rRPC_FWClosePolicyStoreResponse struct {
	PhPolicyStore msfasp.PFW_POLICY_STORE_HANDLE
	Status        ndr.DWORD `ndr:"retval"`
}

// RRPC_FWClosePolicyStore calls RRPC_FWClosePolicyStore (opnum 1) ([MS-FASP] — verify the parameter
// modeling and status handling).
func RRPC_FWClosePolicyStore(rpc ndr.Invoker, phPolicyStore msfasp.PFW_POLICY_STORE_HANDLE) (PhPolicyStore msfasp.PFW_POLICY_STORE_HANDLE, err error) {
	req := &rRPC_FWClosePolicyStoreRequest{
		PhPolicyStore: phPolicyStore,
	}
	var resp rRPC_FWClosePolicyStoreResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRPC_FWClosePolicyStore: %w", err)
		return
	}
	PhPolicyStore = resp.PhPolicyStore
	if uint32(resp.Status) != RemoteFW.StatusSuccess {
		err = fmt.Errorf("RRPC_FWClosePolicyStore failed: %s", RemoteFW.StatusString(uint32(resp.Status)))
	}
	return
}
