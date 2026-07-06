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

// rRPC_FWEnumCryptoSetsRequest carries the [in] parameters of RRPC_FWEnumCryptoSets.
type rRPC_FWEnumCryptoSetsRequest struct {
	HPolicyStore       msfasp.FW_POLICY_STORE_HANDLE
	IpSecPhase         msfasp.FW_IPSEC_PHASE
	DwFilteredByStatus ndr.DWORD
	WFlags             uint16
}

func (*rRPC_FWEnumCryptoSetsRequest) Opnum() uint16 { return RemoteFW.OpnumRRPC_FWEnumCryptoSets }

// rRPC_FWEnumCryptoSetsResponse carries the [out] parameters and return value of RRPC_FWEnumCryptoSets.
type rRPC_FWEnumCryptoSetsResponse struct {
	PdwNumSets   ndr.DWORD
	PpCryptoSets *msfasp.FW_CRYPTO_SET `ndr:"unique"`
	Status       ndr.DWORD             `ndr:"retval"`
}

// RRPC_FWEnumCryptoSets calls RRPC_FWEnumCryptoSets (opnum 26) ([MS-FASP] — verify the parameter
// modeling and status handling).
func RRPC_FWEnumCryptoSets(rpc ndr.Invoker, hPolicyStore msfasp.FW_POLICY_STORE_HANDLE, ipSecPhase msfasp.FW_IPSEC_PHASE, dwFilteredByStatus ndr.DWORD, wFlags uint16) (PdwNumSets ndr.DWORD, PpCryptoSets *msfasp.FW_CRYPTO_SET, err error) {
	req := &rRPC_FWEnumCryptoSetsRequest{
		HPolicyStore:       hPolicyStore,
		IpSecPhase:         ipSecPhase,
		DwFilteredByStatus: dwFilteredByStatus,
		WFlags:             wFlags,
	}
	var resp rRPC_FWEnumCryptoSetsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRPC_FWEnumCryptoSets: %w", err)
		return
	}
	PdwNumSets = resp.PdwNumSets
	PpCryptoSets = resp.PpCryptoSets
	if uint32(resp.Status) != RemoteFW.StatusSuccess {
		err = fmt.Errorf("RRPC_FWEnumCryptoSets failed: %s", RemoteFW.StatusString(uint32(resp.Status)))
	}
	return
}
