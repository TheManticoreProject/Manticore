package functions

import (
	"fmt"

	RemoteFW "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6b5bdd1e-528c-422c-af8c-a4079be4fe48/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msfasp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-fasp"
)

// rRPC_FWDeleteAllCryptoSetsRequest carries the [in] parameters of RRPC_FWDeleteAllCryptoSets.
type rRPC_FWDeleteAllCryptoSetsRequest struct {
	HPolicyStore msfasp.FW_POLICY_STORE_HANDLE
	IpSecPhase   msfasp.FW_IPSEC_PHASE
}

func (*rRPC_FWDeleteAllCryptoSetsRequest) Opnum() uint16 {
	return RemoteFW.OpnumRRPC_FWDeleteAllCryptoSets
}

// rRPC_FWDeleteAllCryptoSetsResponse carries the [out] parameters and return value of RRPC_FWDeleteAllCryptoSets.
type rRPC_FWDeleteAllCryptoSetsResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RRPC_FWDeleteAllCryptoSets calls RRPC_FWDeleteAllCryptoSets (opnum 25) ([MS-FASP] — verify the parameter
// modeling and status handling).
func RRPC_FWDeleteAllCryptoSets(rpc ndr.Invoker, hPolicyStore msfasp.FW_POLICY_STORE_HANDLE, ipSecPhase msfasp.FW_IPSEC_PHASE) (err error) {
	req := &rRPC_FWDeleteAllCryptoSetsRequest{
		HPolicyStore: hPolicyStore,
		IpSecPhase:   ipSecPhase,
	}
	var resp rRPC_FWDeleteAllCryptoSetsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRPC_FWDeleteAllCryptoSets: %w", err)
		return
	}
	if uint32(resp.Status) != RemoteFW.StatusSuccess {
		err = fmt.Errorf("RRPC_FWDeleteAllCryptoSets failed: %s", RemoteFW.StatusString(uint32(resp.Status)))
	}
	return
}
