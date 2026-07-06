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

// frsRpcStartPromotionParentRequest carries the [in] parameters of FrsRpcStartPromotionParent.
type frsRpcStartPromotionParentRequest struct {
	ParentAccount    *ndr.WSTR `ndr:"unique"`
	ParentPassword   *ndr.WSTR `ndr:"unique"`
	ReplicaSetName   *ndr.WSTR `ndr:"unique"`
	ReplicaSetType   *ndr.WSTR `ndr:"unique"`
	CxtionName       *ndr.WSTR `ndr:"unique"`
	PartnerName      *ndr.WSTR `ndr:"unique"`
	PartnerPrincName *ndr.WSTR `ndr:"unique"`
	PartnerAuthLevel ndr.DWORD
	GuidSize         ndr.DWORD
	CxtionGuid       []uint8 `ndr:"unique,size_is=GuidSize"`
	PartnerGuid      []uint8 `ndr:"unique,size_is=GuidSize"`
	ParentGuid       []uint8 `ndr:"unique,size_is=GuidSize"`
}

func (*frsRpcStartPromotionParentRequest) Opnum() uint16 {
	return frsrpc.OpnumFrsRpcStartPromotionParent
}

// frsRpcStartPromotionParentResponse carries the [out] parameters and return value of FrsRpcStartPromotionParent.
type frsRpcStartPromotionParentResponse struct {
	// ParentGuid is the [in, out] buffer returned by the server; its maximum_count is
	// read from the wire, so no size_is sibling is needed on the response.
	ParentGuid []uint8   `ndr:"unique"`
	Status     ndr.DWORD `ndr:"retval"`
}

// FrsRpcStartPromotionParent calls FrsRpcStartPromotionParent (opnum 2) ([MS-FRS1] section 3.3.4.2).
func FrsRpcStartPromotionParent(rpc ndr.Invoker, parentAccount *ndr.WSTR, parentPassword *ndr.WSTR, replicaSetName *ndr.WSTR, replicaSetType *ndr.WSTR, cxtionName *ndr.WSTR, partnerName *ndr.WSTR, partnerPrincName *ndr.WSTR, partnerAuthLevel ndr.DWORD, guidSize ndr.DWORD, cxtionGuid []uint8, partnerGuid []uint8, parentGuid []uint8) (ParentGuid []uint8, err error) {
	req := &frsRpcStartPromotionParentRequest{
		ParentAccount:    parentAccount,
		ParentPassword:   parentPassword,
		ReplicaSetName:   replicaSetName,
		ReplicaSetType:   replicaSetType,
		CxtionName:       cxtionName,
		PartnerName:      partnerName,
		PartnerPrincName: partnerPrincName,
		PartnerAuthLevel: partnerAuthLevel,
		GuidSize:         guidSize,
		CxtionGuid:       cxtionGuid,
		PartnerGuid:      partnerGuid,
		ParentGuid:       parentGuid,
	}
	var resp frsRpcStartPromotionParentResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("FrsRpcStartPromotionParent: %w", err)
		return
	}
	ParentGuid = resp.ParentGuid
	if uint32(resp.Status) != frsrpc.StatusSuccess {
		err = fmt.Errorf("FrsRpcStartPromotionParent failed: %s", frsrpc.StatusString(uint32(resp.Status)))
	}
	return
}
