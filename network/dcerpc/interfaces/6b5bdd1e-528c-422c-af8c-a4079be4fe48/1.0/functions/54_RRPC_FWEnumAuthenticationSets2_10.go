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

// rRPC_FWEnumAuthenticationSets2_10Request carries the [in] parameters of RRPC_FWEnumAuthenticationSets2_10.
type rRPC_FWEnumAuthenticationSets2_10Request struct {
	HPolicyStore       msfasp.FW_POLICY_STORE_HANDLE
	IpSecPhase         msfasp.FW_IPSEC_PHASE
	DwFilteredByStatus ndr.DWORD
	WFlags             uint16
}

func (*rRPC_FWEnumAuthenticationSets2_10Request) Opnum() uint16 {
	return RemoteFW.OpnumRRPC_FWEnumAuthenticationSets2_10
}

// rRPC_FWEnumAuthenticationSets2_10Response carries the [out] parameters and return value of RRPC_FWEnumAuthenticationSets2_10.
type rRPC_FWEnumAuthenticationSets2_10Response struct {
	PdwNumAuthSets ndr.DWORD
	PpAuth         *msfasp.FW_AUTH_SET2_10 `ndr:"unique"`
	Status         ndr.DWORD               `ndr:"retval"`
}

// RRPC_FWEnumAuthenticationSets2_10 calls RRPC_FWEnumAuthenticationSets2_10 (opnum 54) ([MS-FASP] — verify the parameter
// modeling and status handling).
func RRPC_FWEnumAuthenticationSets2_10(rpc ndr.Invoker, hPolicyStore msfasp.FW_POLICY_STORE_HANDLE, ipSecPhase msfasp.FW_IPSEC_PHASE, dwFilteredByStatus ndr.DWORD, wFlags uint16) (PdwNumAuthSets ndr.DWORD, PpAuth *msfasp.FW_AUTH_SET2_10, err error) {
	req := &rRPC_FWEnumAuthenticationSets2_10Request{
		HPolicyStore:       hPolicyStore,
		IpSecPhase:         ipSecPhase,
		DwFilteredByStatus: dwFilteredByStatus,
		WFlags:             wFlags,
	}
	var resp rRPC_FWEnumAuthenticationSets2_10Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRPC_FWEnumAuthenticationSets2_10: %w", err)
		return
	}
	PdwNumAuthSets = resp.PdwNumAuthSets
	PpAuth = resp.PpAuth
	if uint32(resp.Status) != RemoteFW.StatusSuccess {
		err = fmt.Errorf("RRPC_FWEnumAuthenticationSets2_10 failed: %s", RemoteFW.StatusString(uint32(resp.Status)))
	}
	return
}
