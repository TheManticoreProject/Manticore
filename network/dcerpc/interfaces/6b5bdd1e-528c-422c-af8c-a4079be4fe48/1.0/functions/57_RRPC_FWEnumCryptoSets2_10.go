package functions

import (
	"fmt"

	RemoteFW "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6b5bdd1e-528c-422c-af8c-a4079be4fe48/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msfasp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-fasp"
)

// rRPC_FWEnumCryptoSets2_10Request carries the [in] parameters of RRPC_FWEnumCryptoSets2_10.
type rRPC_FWEnumCryptoSets2_10Request struct {
	HPolicyStore       msfasp.FW_POLICY_STORE_HANDLE
	IpSecPhase         msfasp.FW_IPSEC_PHASE
	DwFilteredByStatus ndr.DWORD
	WFlags             uint16
}

func (*rRPC_FWEnumCryptoSets2_10Request) Opnum() uint16 {
	return RemoteFW.OpnumRRPC_FWEnumCryptoSets2_10
}

// rRPC_FWEnumCryptoSets2_10Response carries the [out] parameters and return value of RRPC_FWEnumCryptoSets2_10.
type rRPC_FWEnumCryptoSets2_10Response struct {
	PdwNumSets   ndr.DWORD
	PpCryptoSets *msfasp.FW_CRYPTO_SET `ndr:"unique"`
	Status       ndr.DWORD             `ndr:"retval"`
}

// RRPC_FWEnumCryptoSets2_10 calls RRPC_FWEnumCryptoSets2_10 (opnum 57) ([MS-FASP] — verify the parameter
// modeling and status handling).
func RRPC_FWEnumCryptoSets2_10(rpc ndr.Invoker, hPolicyStore msfasp.FW_POLICY_STORE_HANDLE, ipSecPhase msfasp.FW_IPSEC_PHASE, dwFilteredByStatus ndr.DWORD, wFlags uint16) (PdwNumSets ndr.DWORD, PpCryptoSets *msfasp.FW_CRYPTO_SET, err error) {
	req := &rRPC_FWEnumCryptoSets2_10Request{
		HPolicyStore:       hPolicyStore,
		IpSecPhase:         ipSecPhase,
		DwFilteredByStatus: dwFilteredByStatus,
		WFlags:             wFlags,
	}
	var resp rRPC_FWEnumCryptoSets2_10Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRPC_FWEnumCryptoSets2_10: %w", err)
		return
	}
	PdwNumSets = resp.PdwNumSets
	PpCryptoSets = resp.PpCryptoSets
	if uint32(resp.Status) != RemoteFW.StatusSuccess {
		err = fmt.Errorf("RRPC_FWEnumCryptoSets2_10 failed: %s", RemoteFW.StatusString(uint32(resp.Status)))
	}
	return
}
