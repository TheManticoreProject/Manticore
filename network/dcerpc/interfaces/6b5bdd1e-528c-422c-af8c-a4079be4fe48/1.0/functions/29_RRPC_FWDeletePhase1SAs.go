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

// rRPC_FWDeletePhase1SAsRequest carries the [in] parameters of RRPC_FWDeletePhase1SAs.
type rRPC_FWDeletePhase1SAsRequest struct {
	HPolicyStore msfasp.FW_POLICY_STORE_HANDLE
	PEndpoints   *msfasp.FW_ENDPOINTS `ndr:"unique"`
}

func (*rRPC_FWDeletePhase1SAsRequest) Opnum() uint16 { return RemoteFW.OpnumRRPC_FWDeletePhase1SAs }

// rRPC_FWDeletePhase1SAsResponse carries the [out] parameters and return value of RRPC_FWDeletePhase1SAs.
type rRPC_FWDeletePhase1SAsResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RRPC_FWDeletePhase1SAs calls RRPC_FWDeletePhase1SAs (opnum 29) ([MS-FASP] — verify the parameter
// modeling and status handling).
func RRPC_FWDeletePhase1SAs(rpc ndr.Invoker, hPolicyStore msfasp.FW_POLICY_STORE_HANDLE, pEndpoints *msfasp.FW_ENDPOINTS) (err error) {
	req := &rRPC_FWDeletePhase1SAsRequest{
		HPolicyStore: hPolicyStore,
		PEndpoints:   pEndpoints,
	}
	var resp rRPC_FWDeletePhase1SAsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRPC_FWDeletePhase1SAs: %w", err)
		return
	}
	if uint32(resp.Status) != RemoteFW.StatusSuccess {
		err = fmt.Errorf("RRPC_FWDeletePhase1SAs failed: %s", RemoteFW.StatusString(uint32(resp.Status)))
	}
	return
}
