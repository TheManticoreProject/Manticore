package functions

// IDL source: [MS-FASP] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-fasp/1503b9d7-7fec-4793-9972-6ad58720c9db
// A fetched copy is kept at ms-fasp.idl in the interface directory.

import (
	"fmt"

	RemoteFW "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6b5bdd1e-528c-422c-af8c-a4079be4fe48/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msfasp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-fasp"
)

// rRPC_FWAddAuthenticationSetRequest carries the [in] parameters of RRPC_FWAddAuthenticationSet.
type rRPC_FWAddAuthenticationSetRequest struct {
	HPolicyStore msfasp.FW_POLICY_STORE_HANDLE
	PAuth        msfasp.FW_AUTH_SET2_10
}

func (*rRPC_FWAddAuthenticationSetRequest) Opnum() uint16 {
	return RemoteFW.OpnumRRPC_FWAddAuthenticationSet
}

// rRPC_FWAddAuthenticationSetResponse carries the [out] parameters and return value of RRPC_FWAddAuthenticationSet.
type rRPC_FWAddAuthenticationSetResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RRPC_FWAddAuthenticationSet calls RRPC_FWAddAuthenticationSet (opnum 17) ([MS-FASP] — verify the parameter
// modeling and status handling).
func RRPC_FWAddAuthenticationSet(rpc ndr.Invoker, hPolicyStore msfasp.FW_POLICY_STORE_HANDLE, pAuth msfasp.FW_AUTH_SET2_10) (err error) {
	req := &rRPC_FWAddAuthenticationSetRequest{
		HPolicyStore: hPolicyStore,
		PAuth:        pAuth,
	}
	var resp rRPC_FWAddAuthenticationSetResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRPC_FWAddAuthenticationSet: %w", err)
		return
	}
	if uint32(resp.Status) != RemoteFW.StatusSuccess {
		err = fmt.Errorf("RRPC_FWAddAuthenticationSet failed: %s", RemoteFW.StatusString(uint32(resp.Status)))
	}
	return
}
