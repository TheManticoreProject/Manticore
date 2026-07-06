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

// rRPC_FWDeletePhase2SAsRequest carries the [in] parameters of RRPC_FWDeletePhase2SAs.
type rRPC_FWDeletePhase2SAsRequest struct {
	HPolicyStore msfasp.FW_POLICY_STORE_HANDLE
	PEndpoints   *msfasp.FW_ENDPOINTS `ndr:"unique"`
}

func (*rRPC_FWDeletePhase2SAsRequest) Opnum() uint16 { return RemoteFW.OpnumRRPC_FWDeletePhase2SAs }

// rRPC_FWDeletePhase2SAsResponse carries the [out] parameters and return value of RRPC_FWDeletePhase2SAs.
type rRPC_FWDeletePhase2SAsResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// RRPC_FWDeletePhase2SAs calls RRPC_FWDeletePhase2SAs (opnum 30) ([MS-FASP] — verify the parameter
// modeling and status handling).
func RRPC_FWDeletePhase2SAs(rpc ndr.Invoker, hPolicyStore msfasp.FW_POLICY_STORE_HANDLE, pEndpoints *msfasp.FW_ENDPOINTS) (err error) {
	req := &rRPC_FWDeletePhase2SAsRequest{
		HPolicyStore: hPolicyStore,
		PEndpoints:   pEndpoints,
	}
	var resp rRPC_FWDeletePhase2SAsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRPC_FWDeletePhase2SAs: %w", err)
		return
	}
	if uint32(resp.Status) != RemoteFW.StatusSuccess {
		err = fmt.Errorf("RRPC_FWDeletePhase2SAs failed: %s", RemoteFW.StatusString(uint32(resp.Status)))
	}
	return
}
