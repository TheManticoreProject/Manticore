package functions

import (
	"fmt"

	RemoteFW "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6b5bdd1e-528c-422c-af8c-a4079be4fe48/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msfasp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-fasp"
)

// rRPC_FWAddAuthenticationSet2_10Request carries the [in] parameters of RRPC_FWAddAuthenticationSet2_10.
type rRPC_FWAddAuthenticationSet2_10Request struct {
	HPolicyStore msfasp.FW_POLICY_STORE_HANDLE
	PAuth        msfasp.FW_AUTH_SET2_10
}

func (*rRPC_FWAddAuthenticationSet2_10Request) Opnum() uint16 {
	return RemoteFW.OpnumRRPC_FWAddAuthenticationSet2_10
}

// rRPC_FWAddAuthenticationSet2_10Response carries the [out] parameters and return value of RRPC_FWAddAuthenticationSet2_10.
type rRPC_FWAddAuthenticationSet2_10Response struct {
	PStatus msfasp.FW_RULE_STATUS
	Status  ndr.DWORD `ndr:"retval"`
}

// RRPC_FWAddAuthenticationSet2_10 calls RRPC_FWAddAuthenticationSet2_10 (opnum 52) ([MS-FASP] — verify the parameter
// modeling and status handling).
func RRPC_FWAddAuthenticationSet2_10(rpc ndr.Invoker, hPolicyStore msfasp.FW_POLICY_STORE_HANDLE, pAuth msfasp.FW_AUTH_SET2_10) (PStatus msfasp.FW_RULE_STATUS, err error) {
	req := &rRPC_FWAddAuthenticationSet2_10Request{
		HPolicyStore: hPolicyStore,
		PAuth:        pAuth,
	}
	var resp rRPC_FWAddAuthenticationSet2_10Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRPC_FWAddAuthenticationSet2_10: %w", err)
		return
	}
	PStatus = resp.PStatus
	if uint32(resp.Status) != RemoteFW.StatusSuccess {
		err = fmt.Errorf("RRPC_FWAddAuthenticationSet2_10 failed: %s", RemoteFW.StatusString(uint32(resp.Status)))
	}
	return
}
