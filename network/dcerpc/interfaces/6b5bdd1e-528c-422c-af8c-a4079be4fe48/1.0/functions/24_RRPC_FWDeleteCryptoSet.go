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

// rRPC_FWDeleteCryptoSetRequest carries the [in] parameters of RRPC_FWDeleteCryptoSet.
type rRPC_FWDeleteCryptoSetRequest struct {
	HPolicyStore msfasp.FW_POLICY_STORE_HANDLE
	IpSecPhase   msfasp.FW_IPSEC_PHASE
	WszSetId     ndr.WSTR
}

func (*rRPC_FWDeleteCryptoSetRequest) Opnum() uint16 { return RemoteFW.OpnumRRPC_FWDeleteCryptoSet }

// rRPC_FWDeleteCryptoSetResponse carries the [out] parameters and return value of RRPC_FWDeleteCryptoSet.
type rRPC_FWDeleteCryptoSetResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RRPC_FWDeleteCryptoSet calls RRPC_FWDeleteCryptoSet (opnum 24) ([MS-FASP] — verify the parameter
// modeling and status handling).
func RRPC_FWDeleteCryptoSet(rpc ndr.Invoker, hPolicyStore msfasp.FW_POLICY_STORE_HANDLE, ipSecPhase msfasp.FW_IPSEC_PHASE, wszSetId ndr.WSTR) (err error) {
	req := &rRPC_FWDeleteCryptoSetRequest{
		HPolicyStore: hPolicyStore,
		IpSecPhase:   ipSecPhase,
		WszSetId:     wszSetId,
	}
	var resp rRPC_FWDeleteCryptoSetResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRPC_FWDeleteCryptoSet: %w", err)
		return
	}
	if uint32(resp.Status) != RemoteFW.StatusSuccess {
		err = fmt.Errorf("RRPC_FWDeleteCryptoSet failed: %s", RemoteFW.StatusString(uint32(resp.Status)))
	}
	return
}
