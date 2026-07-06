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

// rRPC_FWDeleteAllConnectionSecurityRulesRequest carries the [in] parameters of RRPC_FWDeleteAllConnectionSecurityRules.
type rRPC_FWDeleteAllConnectionSecurityRulesRequest struct {
	HPolicyStore msfasp.FW_POLICY_STORE_HANDLE
}

func (*rRPC_FWDeleteAllConnectionSecurityRulesRequest) Opnum() uint16 {
	return RemoteFW.OpnumRRPC_FWDeleteAllConnectionSecurityRules
}

// rRPC_FWDeleteAllConnectionSecurityRulesResponse carries the [out] parameters and return value of RRPC_FWDeleteAllConnectionSecurityRules.
type rRPC_FWDeleteAllConnectionSecurityRulesResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RRPC_FWDeleteAllConnectionSecurityRules calls RRPC_FWDeleteAllConnectionSecurityRules (opnum 15) ([MS-FASP] — verify the parameter
// modeling and status handling).
func RRPC_FWDeleteAllConnectionSecurityRules(rpc ndr.Invoker, hPolicyStore msfasp.FW_POLICY_STORE_HANDLE) (err error) {
	req := &rRPC_FWDeleteAllConnectionSecurityRulesRequest{
		HPolicyStore: hPolicyStore,
	}
	var resp rRPC_FWDeleteAllConnectionSecurityRulesResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRPC_FWDeleteAllConnectionSecurityRules: %w", err)
		return
	}
	if uint32(resp.Status) != RemoteFW.StatusSuccess {
		err = fmt.Errorf("RRPC_FWDeleteAllConnectionSecurityRules failed: %s", RemoteFW.StatusString(uint32(resp.Status)))
	}
	return
}
