package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0/structures"
)

// lsarSetForestTrustInformationRequest is the [in] parameter set of
// LsarSetForestTrustInformation: an open policy handle, the inline [ref] name of the
// trusted domain, the highest forest-trust record type the caller understands, the inline
// [ref] forest-trust information to set, and the CheckOnly flag (when non-zero the server
// validates the information and reports collisions without storing it).
type lsarSetForestTrustInformationRequest struct {
	PolicyHandle      structures.LSAPR_HANDLE
	TrustedDomainName dtyp.RPC_UNICODE_STRING
	HighestRecordType structures.LSA_FOREST_TRUST_RECORD_TYPE
	ForestTrustInfo   structures.LSA_FOREST_TRUST_INFORMATION
	CheckOnly         uint8
}

func (*lsarSetForestTrustInformationRequest) Opnum() uint16 {
	return lsarpc.OpnumLsarSetForestTrustInformation
}

// lsarSetForestTrustInformationResponse is the reply: the [unique] collision information
// describing any conflicts with existing trusts followed by the NTSTATUS return value.
type lsarSetForestTrustInformationResponse struct {
	CollisionInfo *structures.LSA_FOREST_TRUST_COLLISION_INFORMATION `ndr:"unique"`
	Status        ndr.DWORD
}

// LsarSetForestTrustInformation calls LsarSetForestTrustInformation (opnum 74), establishing
// the forest-trust information for the named trusted domain ([MS-LSAD] 3.1.4.7.16). When
// checkOnly is true the server only validates the information; the returned pointer holds
// any collision information and may be nil when there are no collisions.
func LsarSetForestTrustInformation(rpc *client.Client, policyHandle structures.LSAPR_HANDLE, trustedDomainName string, highestRecordType structures.LSA_FOREST_TRUST_RECORD_TYPE, forestTrustInfo structures.LSA_FOREST_TRUST_INFORMATION, checkOnly uint8) (*structures.LSA_FOREST_TRUST_COLLISION_INFORMATION, error) {
	req := &lsarSetForestTrustInformationRequest{
		PolicyHandle:      policyHandle,
		TrustedDomainName: dtyp.NewUnicodeString(trustedDomainName),
		HighestRecordType: highestRecordType,
		ForestTrustInfo:   forestTrustInfo,
		CheckOnly:         checkOnly,
	}
	var resp lsarSetForestTrustInformationResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return nil, fmt.Errorf("LsarSetForestTrustInformation: %w", err)
	}
	if uint32(resp.Status) != lsarpc.StatusSuccess {
		return resp.CollisionInfo, fmt.Errorf("LsarSetForestTrustInformation failed: %s", lsarpc.StatusString(uint32(resp.Status)))
	}
	return resp.CollisionInfo, nil
}
