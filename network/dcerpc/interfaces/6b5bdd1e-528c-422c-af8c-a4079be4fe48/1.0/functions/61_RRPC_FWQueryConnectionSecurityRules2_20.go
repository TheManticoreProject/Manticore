package functions

import (
	"fmt"

	RemoteFW "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6b5bdd1e-528c-422c-af8c-a4079be4fe48/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msfasp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-fasp"
)

// rRPC_FWQueryConnectionSecurityRules2_20Request carries the [in] parameters of RRPC_FWQueryConnectionSecurityRules2_20.
type rRPC_FWQueryConnectionSecurityRules2_20Request struct {
	HPolicyStore msfasp.FW_POLICY_STORE_HANDLE
	PQuery       msfasp.FW_QUERY
	WFlags       uint16
}

func (*rRPC_FWQueryConnectionSecurityRules2_20Request) Opnum() uint16 {
	return RemoteFW.OpnumRRPC_FWQueryConnectionSecurityRules2_20
}

// rRPC_FWQueryConnectionSecurityRules2_20Response carries the [out] parameters and return value of RRPC_FWQueryConnectionSecurityRules2_20.
type rRPC_FWQueryConnectionSecurityRules2_20Response struct {
	PdwNumRules ndr.DWORD
	PpRules     *msfasp.FW_CS_RULE `ndr:"unique"`
	Status      ndr.DWORD          `ndr:"retval"`
}

// RRPC_FWQueryConnectionSecurityRules2_20 calls RRPC_FWQueryConnectionSecurityRules2_20 (opnum 61) ([MS-FASP] — verify the parameter
// modeling and status handling).
func RRPC_FWQueryConnectionSecurityRules2_20(rpc ndr.Invoker, hPolicyStore msfasp.FW_POLICY_STORE_HANDLE, pQuery msfasp.FW_QUERY, wFlags uint16) (PdwNumRules ndr.DWORD, PpRules *msfasp.FW_CS_RULE, err error) {
	req := &rRPC_FWQueryConnectionSecurityRules2_20Request{
		HPolicyStore: hPolicyStore,
		PQuery:       pQuery,
		WFlags:       wFlags,
	}
	var resp rRPC_FWQueryConnectionSecurityRules2_20Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRPC_FWQueryConnectionSecurityRules2_20: %w", err)
		return
	}
	PdwNumRules = resp.PdwNumRules
	PpRules = resp.PpRules
	if uint32(resp.Status) != RemoteFW.StatusSuccess {
		err = fmt.Errorf("RRPC_FWQueryConnectionSecurityRules2_20 failed: %s", RemoteFW.StatusString(uint32(resp.Status)))
	}
	return
}
