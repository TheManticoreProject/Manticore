package functions

// IDL source: [MS-FRS1] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-frs1/dd60a0d9-176a-46f4-9904-000172041b92
// A fetched copy is kept at ms-frs1.idl in the interface directory.

import (
	"fmt"

	frsapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/d049b186-814f-11d1-9a3c-00c04fc9b232/1.1"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// ntFrsApi_Rpc_ForceReplicationRequest carries the [in] parameters of NtFrsApi_Rpc_ForceReplication.
type ntFrsApi_Rpc_ForceReplicationRequest struct {
	ReplicaSetGuid *msdtyp.GUID `ndr:"unique"`
	CxtionGuid     *msdtyp.GUID `ndr:"unique"`
	ReplicaSetName *ndr.WSTR    `ndr:"unique"`
	PartnerDnsName *ndr.WSTR    `ndr:"unique"`
}

func (*ntFrsApi_Rpc_ForceReplicationRequest) Opnum() uint16 {
	return frsapi.OpnumNtFrsApi_Rpc_ForceReplication
}

// NtFrsApi_Rpc_ForceReplication calls NtFrsApi_Rpc_ForceReplication (opnum 10) ([MS-FRS1] section 3.2.4.6).
func NtFrsApi_Rpc_ForceReplication(rpc ndr.Invoker, replicaSetGuid *msdtyp.GUID, cxtionGuid *msdtyp.GUID, replicaSetName *ndr.WSTR, partnerDnsName *ndr.WSTR) (err error) {
	req := &ntFrsApi_Rpc_ForceReplicationRequest{
		ReplicaSetGuid: replicaSetGuid,
		CxtionGuid:     cxtionGuid,
		ReplicaSetName: replicaSetName,
		PartnerDnsName: partnerDnsName,
	}
	var resp statusResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NtFrsApi_Rpc_ForceReplication: %w", err)
		return
	}
	if uint32(resp.Status) != frsapi.StatusSuccess {
		err = fmt.Errorf("NtFrsApi_Rpc_ForceReplication failed: %s", frsapi.StatusString(uint32(resp.Status)))
	}
	return
}
