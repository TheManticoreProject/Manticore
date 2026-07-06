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

// rRPC_FWAddAuthenticationSet2_20Request carries the [in] parameters of RRPC_FWAddAuthenticationSet2_20.
type rRPC_FWAddAuthenticationSet2_20Request struct {
	HPolicyStore msfasp.FW_POLICY_STORE_HANDLE
	PAuth        msfasp.FW_AUTH_SET
}

func (*rRPC_FWAddAuthenticationSet2_20Request) Opnum() uint16 {
	return RemoteFW.OpnumRRPC_FWAddAuthenticationSet2_20
}

// rRPC_FWAddAuthenticationSet2_20Response carries the [out] parameters and return value of RRPC_FWAddAuthenticationSet2_20.
type rRPC_FWAddAuthenticationSet2_20Response struct {
	PStatus msfasp.FW_RULE_STATUS
	Status  ndr.DWORD `ndr:"retval"`
}

// RRPC_FWAddAuthenticationSet2_20 calls RRPC_FWAddAuthenticationSet2_20 (opnum 62) ([MS-FASP] — verify the parameter
// modeling and status handling).
func RRPC_FWAddAuthenticationSet2_20(rpc ndr.Invoker, hPolicyStore msfasp.FW_POLICY_STORE_HANDLE, pAuth msfasp.FW_AUTH_SET) (PStatus msfasp.FW_RULE_STATUS, err error) {
	req := &rRPC_FWAddAuthenticationSet2_20Request{
		HPolicyStore: hPolicyStore,
		PAuth:        pAuth,
	}
	var resp rRPC_FWAddAuthenticationSet2_20Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRPC_FWAddAuthenticationSet2_20: %w", err)
		return
	}
	PStatus = resp.PStatus
	if uint32(resp.Status) != RemoteFW.StatusSuccess {
		err = fmt.Errorf("RRPC_FWAddAuthenticationSet2_20 failed: %s", RemoteFW.StatusString(uint32(resp.Status)))
	}
	return
}
