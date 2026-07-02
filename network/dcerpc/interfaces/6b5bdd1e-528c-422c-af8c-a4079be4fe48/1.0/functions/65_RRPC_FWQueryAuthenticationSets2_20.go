package functions

import (
	"fmt"

	RemoteFW "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6b5bdd1e-528c-422c-af8c-a4079be4fe48/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msfasp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-fasp"
)

// rRPC_FWQueryAuthenticationSets2_20Request carries the [in] parameters of RRPC_FWQueryAuthenticationSets2_20.
type rRPC_FWQueryAuthenticationSets2_20Request struct {
	HPolicyStore msfasp.FW_POLICY_STORE_HANDLE
	IPsecPhase   msfasp.FW_IPSEC_PHASE
	PQuery       msfasp.FW_QUERY
	WFlags       uint16
}

func (*rRPC_FWQueryAuthenticationSets2_20Request) Opnum() uint16 {
	return RemoteFW.OpnumRRPC_FWQueryAuthenticationSets2_20
}

// rRPC_FWQueryAuthenticationSets2_20Response carries the [out] parameters and return value of RRPC_FWQueryAuthenticationSets2_20.
type rRPC_FWQueryAuthenticationSets2_20Response struct {
	PdwNumSets ndr.DWORD
	PpAuthSets *msfasp.FW_AUTH_SET `ndr:"unique"`
	Status     ndr.DWORD           `ndr:"retval"`
}

// RRPC_FWQueryAuthenticationSets2_20 calls RRPC_FWQueryAuthenticationSets2_20 (opnum 65) ([MS-FASP] — verify the parameter
// modeling and status handling).
func RRPC_FWQueryAuthenticationSets2_20(rpc ndr.Invoker, hPolicyStore msfasp.FW_POLICY_STORE_HANDLE, iPsecPhase msfasp.FW_IPSEC_PHASE, pQuery msfasp.FW_QUERY, wFlags uint16) (PdwNumSets ndr.DWORD, PpAuthSets *msfasp.FW_AUTH_SET, err error) {
	req := &rRPC_FWQueryAuthenticationSets2_20Request{
		HPolicyStore: hPolicyStore,
		IPsecPhase:   iPsecPhase,
		PQuery:       pQuery,
		WFlags:       wFlags,
	}
	var resp rRPC_FWQueryAuthenticationSets2_20Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRPC_FWQueryAuthenticationSets2_20: %w", err)
		return
	}
	PdwNumSets = resp.PdwNumSets
	PpAuthSets = resp.PpAuthSets
	if uint32(resp.Status) != RemoteFW.StatusSuccess {
		err = fmt.Errorf("RRPC_FWQueryAuthenticationSets2_20 failed: %s", RemoteFW.StatusString(uint32(resp.Status)))
	}
	return
}
