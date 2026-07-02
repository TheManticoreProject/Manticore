package functions

import (
	"fmt"

	RemoteFW "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6b5bdd1e-528c-422c-af8c-a4079be4fe48/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msfasp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-fasp"
)

// rRPC_FWAddCryptoSet2_10Request carries the [in] parameters of RRPC_FWAddCryptoSet2_10.
type rRPC_FWAddCryptoSet2_10Request struct {
	HPolicyStore msfasp.FW_POLICY_STORE_HANDLE
	PCrypto      msfasp.FW_CRYPTO_SET
}

func (*rRPC_FWAddCryptoSet2_10Request) Opnum() uint16 { return RemoteFW.OpnumRRPC_FWAddCryptoSet2_10 }

// rRPC_FWAddCryptoSet2_10Response carries the [out] parameters and return value of RRPC_FWAddCryptoSet2_10.
type rRPC_FWAddCryptoSet2_10Response struct {
	PStatus msfasp.FW_RULE_STATUS
	Status  ndr.DWORD `ndr:"retval"`
}

// RRPC_FWAddCryptoSet2_10 calls RRPC_FWAddCryptoSet2_10 (opnum 55) ([MS-FASP] — verify the parameter
// modeling and status handling).
func RRPC_FWAddCryptoSet2_10(rpc ndr.Invoker, hPolicyStore msfasp.FW_POLICY_STORE_HANDLE, pCrypto msfasp.FW_CRYPTO_SET) (PStatus msfasp.FW_RULE_STATUS, err error) {
	req := &rRPC_FWAddCryptoSet2_10Request{
		HPolicyStore: hPolicyStore,
		PCrypto:      pCrypto,
	}
	var resp rRPC_FWAddCryptoSet2_10Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRPC_FWAddCryptoSet2_10: %w", err)
		return
	}
	PStatus = resp.PStatus
	if uint32(resp.Status) != RemoteFW.StatusSuccess {
		err = fmt.Errorf("RRPC_FWAddCryptoSet2_10 failed: %s", RemoteFW.StatusString(uint32(resp.Status)))
	}
	return
}
