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

// rRPC_FWOpenPolicyStoreRequest carries the [in] parameters of RRPC_FWOpenPolicyStore.
type rRPC_FWOpenPolicyStoreRequest struct {
	BinaryVersion uint16
	StoreType     msfasp.FW_STORE_TYPE
	AccessRight   msfasp.FW_POLICY_ACCESS_RIGHT
	DwFlags       ndr.DWORD
}

func (*rRPC_FWOpenPolicyStoreRequest) Opnum() uint16 { return RemoteFW.OpnumRRPC_FWOpenPolicyStore }

// rRPC_FWOpenPolicyStoreResponse carries the [out] parameters and return value of RRPC_FWOpenPolicyStore.
type rRPC_FWOpenPolicyStoreResponse struct {
	PhPolicyStore msfasp.PFW_POLICY_STORE_HANDLE
	Status        ndr.DWORD `ndr:"retval"`
}

// RRPC_FWOpenPolicyStore calls RRPC_FWOpenPolicyStore (opnum 0) ([MS-FASP] — verify the parameter
// modeling and status handling).
func RRPC_FWOpenPolicyStore(rpc ndr.Invoker, binaryVersion uint16, storeType msfasp.FW_STORE_TYPE, accessRight msfasp.FW_POLICY_ACCESS_RIGHT, dwFlags ndr.DWORD) (PhPolicyStore msfasp.PFW_POLICY_STORE_HANDLE, err error) {
	req := &rRPC_FWOpenPolicyStoreRequest{
		BinaryVersion: binaryVersion,
		StoreType:     storeType,
		AccessRight:   accessRight,
		DwFlags:       dwFlags,
	}
	var resp rRPC_FWOpenPolicyStoreResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RRPC_FWOpenPolicyStore: %w", err)
		return
	}
	PhPolicyStore = resp.PhPolicyStore
	if uint32(resp.Status) != RemoteFW.StatusSuccess {
		err = fmt.Errorf("RRPC_FWOpenPolicyStore failed: %s", RemoteFW.StatusString(uint32(resp.Status)))
	}
	return
}
