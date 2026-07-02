package functions

import (
	"fmt"

	RemoteFW "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6b5bdd1e-528c-422c-af8c-a4079be4fe48/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msfasp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-fasp"
)

// rRPC_FWSetCryptoSetRequest carries the [in] parameters of RRPC_FWSetCryptoSet.
type rRPC_FWSetCryptoSetRequest struct {
	HPolicyStore msfasp.FW_POLICY_STORE_HANDLE
	PCrypto      msfasp.FW_CRYPTO_SET
}

func (*rRPC_FWSetCryptoSetRequest) Opnum() uint16 { return RemoteFW.OpnumRRPC_FWSetCryptoSet }

// rRPC_FWSetCryptoSetResponse carries the [out] parameters and return value of RRPC_FWSetCryptoSet.
type rRPC_FWSetCryptoSetResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RRPC_FWSetCryptoSet calls RRPC_FWSetCryptoSet (opnum 23) ([MS-FASP] — verify the parameter
// modeling and status handling).
func RRPC_FWSetCryptoSet(rpc ndr.Invoker, hPolicyStore msfasp.FW_POLICY_STORE_HANDLE, pCrypto msfasp.FW_CRYPTO_SET) (err error) {
	req := &rRPC_FWSetCryptoSetRequest{
		HPolicyStore: hPolicyStore,
		PCrypto:      pCrypto,
	}
	var resp rRPC_FWSetCryptoSetResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRPC_FWSetCryptoSet: %w", err)
		return
	}
	if uint32(resp.Status) != RemoteFW.StatusSuccess {
		err = fmt.Errorf("RRPC_FWSetCryptoSet failed: %s", RemoteFW.StatusString(uint32(resp.Status)))
	}
	return
}
