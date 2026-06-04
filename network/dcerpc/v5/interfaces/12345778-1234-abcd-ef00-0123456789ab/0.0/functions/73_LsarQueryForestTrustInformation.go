package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0/structures"
)

// lsarQueryForestTrustInformationRequest is the [in] parameter set of
// LsarQueryForestTrustInformation: an open policy handle, the inline [ref] name of the
// trusted domain to query, and the highest forest-trust record type the caller understands.
type lsarQueryForestTrustInformationRequest struct {
	PolicyHandle      structures.LSAPR_HANDLE
	TrustedDomainName dtyp.RPC_UNICODE_STRING
	HighestRecordType structures.LSA_FOREST_TRUST_RECORD_TYPE
}

func (*lsarQueryForestTrustInformationRequest) Opnum() uint16 {
	return lsarpc.OpnumLsarQueryForestTrustInformation
}

// lsarQueryForestTrustInformationResponse is the reply: the [unique] forest-trust
// information returned by the server followed by the NTSTATUS return value.
type lsarQueryForestTrustInformationResponse struct {
	ForestTrustInfo *structures.LSA_FOREST_TRUST_INFORMATION `ndr:"unique"`
	Status          ndr.DWORD                                `ndr:"retval"`
}

// LsarQueryForestTrustInformation calls LsarQueryForestTrustInformation (opnum 73) and
// returns the forest-trust information for the named trusted domain ([MS-LSAD] 3.1.4.7.18).
// The returned pointer may be nil if the server did not return any information.
func LsarQueryForestTrustInformation(rpc *client.Client, policyHandle structures.LSAPR_HANDLE, trustedDomainName string, highestRecordType structures.LSA_FOREST_TRUST_RECORD_TYPE) (*structures.LSA_FOREST_TRUST_INFORMATION, error) {
	req := &lsarQueryForestTrustInformationRequest{
		PolicyHandle:      policyHandle,
		TrustedDomainName: dtyp.NewUnicodeString(trustedDomainName),
		HighestRecordType: highestRecordType,
	}
	var resp lsarQueryForestTrustInformationResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return nil, fmt.Errorf("LsarQueryForestTrustInformation: %w", err)
	}
	if uint32(resp.Status) != lsarpc.StatusSuccess {
		return resp.ForestTrustInfo, fmt.Errorf("LsarQueryForestTrustInformation failed: %s", lsarpc.StatusString(uint32(resp.Status)))
	}
	return resp.ForestTrustInfo, nil
}
