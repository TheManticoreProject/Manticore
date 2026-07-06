package functions

// IDL source: [MS-FRS1] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-frs1/dd60a0d9-176a-46f4-9904-000172041b92
// A fetched copy is kept at ms-frs1.idl in the interface directory.

import (
	"fmt"

	frsrpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/f5cc59b4-4264-101a-8c59-08002b2f8426/1.1"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// frsRpcVerifyPromotionParentRequest carries the [in] parameters of FrsRpcVerifyPromotionParent.
type frsRpcVerifyPromotionParentRequest struct {
	ParentAccount    *ndr.WSTR `ndr:"unique"`
	ParentPassword   *ndr.WSTR `ndr:"unique"`
	ReplicaSetName   *ndr.WSTR `ndr:"unique"`
	ReplicaSetType   *ndr.WSTR `ndr:"unique"`
	PartnerAuthLevel ndr.DWORD
	GuidSize         ndr.DWORD
}

func (*frsRpcVerifyPromotionParentRequest) Opnum() uint16 {
	return frsrpc.OpnumFrsRpcVerifyPromotionParent
}

// FrsRpcVerifyPromotionParent calls FrsRpcVerifyPromotionParent (opnum 1) ([MS-FRS1] section 3.3.4.5).
func FrsRpcVerifyPromotionParent(rpc ndr.Invoker, parentAccount *ndr.WSTR, parentPassword *ndr.WSTR, replicaSetName *ndr.WSTR, replicaSetType *ndr.WSTR, partnerAuthLevel ndr.DWORD, guidSize ndr.DWORD) (err error) {
	req := &frsRpcVerifyPromotionParentRequest{
		ParentAccount:    parentAccount,
		ParentPassword:   parentPassword,
		ReplicaSetName:   replicaSetName,
		ReplicaSetType:   replicaSetType,
		PartnerAuthLevel: partnerAuthLevel,
		GuidSize:         guidSize,
	}
	var resp statusResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FrsRpcVerifyPromotionParent: %w", err)
		return
	}
	if uint32(resp.Status) != frsrpc.StatusSuccess {
		err = fmt.Errorf("FrsRpcVerifyPromotionParent failed: %s", frsrpc.StatusString(uint32(resp.Status)))
	}
	return
}
