package functions

// IDL source: [MS-FRS2] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-frs2/39bd498b-2a94-41b7-914e-10921d543912
// A fetched copy is kept at ms-frs2.idl in the interface directory.

import (
	"fmt"

	FrsTransport "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/897e2e5f-93f3-4376-9c9c-fd2277495c27/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msfrs2 "github.com/TheManticoreProject/Manticore/windows/protocols/ms-frs2"
)

// checkConnectivityRequest carries the [in] parameters of CheckConnectivity.
type checkConnectivityRequest struct {
	ReplicaSetId msfrs2.FRS_REPLICA_SET_ID
	ConnectionId msfrs2.FRS_CONNECTION_ID
}

func (*checkConnectivityRequest) Opnum() uint16 { return FrsTransport.OpnumCheckConnectivity }

// checkConnectivityResponse carries the [out] parameters and return value of CheckConnectivity.
type checkConnectivityResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// CheckConnectivity calls CheckConnectivity (opnum 0) ([MS-FRS2] — verify the parameter
// modeling and status handling).
func CheckConnectivity(rpc ndr.Invoker, replicaSetId msfrs2.FRS_REPLICA_SET_ID, connectionId msfrs2.FRS_CONNECTION_ID) (err error) {
	req := &checkConnectivityRequest{
		ReplicaSetId: replicaSetId,
		ConnectionId: connectionId,
	}
	var resp checkConnectivityResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("CheckConnectivity: %w", err)
		return
	}
	if uint32(resp.Status) != FrsTransport.StatusSuccess {
		err = fmt.Errorf("CheckConnectivity failed: %s", FrsTransport.StatusString(uint32(resp.Status)))
	}
	return
}
