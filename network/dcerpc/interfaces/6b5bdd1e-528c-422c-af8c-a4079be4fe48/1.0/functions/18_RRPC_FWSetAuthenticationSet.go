package functions

import (
	"fmt"

	RemoteFW "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6b5bdd1e-528c-422c-af8c-a4079be4fe48/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msfasp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-fasp"
)

// rRPC_FWSetAuthenticationSetRequest carries the [in] parameters of RRPC_FWSetAuthenticationSet.
type rRPC_FWSetAuthenticationSetRequest struct {
	HPolicyStore msfasp.FW_POLICY_STORE_HANDLE
	PAuth        msfasp.FW_AUTH_SET2_10
}

func (*rRPC_FWSetAuthenticationSetRequest) Opnum() uint16 {
	return RemoteFW.OpnumRRPC_FWSetAuthenticationSet
}

// rRPC_FWSetAuthenticationSetResponse carries the [out] parameters and return value of RRPC_FWSetAuthenticationSet.
type rRPC_FWSetAuthenticationSetResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RRPC_FWSetAuthenticationSet calls RRPC_FWSetAuthenticationSet (opnum 18) ([MS-FASP] — verify the parameter
// modeling and status handling).
func RRPC_FWSetAuthenticationSet(rpc ndr.Invoker, hPolicyStore msfasp.FW_POLICY_STORE_HANDLE, pAuth msfasp.FW_AUTH_SET2_10) (err error) {
	req := &rRPC_FWSetAuthenticationSetRequest{
		HPolicyStore: hPolicyStore,
		PAuth:        pAuth,
	}
	var resp rRPC_FWSetAuthenticationSetResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRPC_FWSetAuthenticationSet: %w", err)
		return
	}
	if uint32(resp.Status) != RemoteFW.StatusSuccess {
		err = fmt.Errorf("RRPC_FWSetAuthenticationSet failed: %s", RemoteFW.StatusString(uint32(resp.Status)))
	}
	return
}
