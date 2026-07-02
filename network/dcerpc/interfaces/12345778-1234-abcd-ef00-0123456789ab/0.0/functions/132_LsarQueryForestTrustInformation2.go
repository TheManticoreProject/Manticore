package functions

import (
	"fmt"

	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mslsad "github.com/TheManticoreProject/Manticore/windows/protocols/ms-lsad"
)

// lsarQueryForestTrustInformation2Request carries the [in] parameters of LsarQueryForestTrustInformation2.
type lsarQueryForestTrustInformation2Request struct {
	PolicyHandle      mslsad.LSAPR_HANDLE
	TrustedDomainName mslsad.LSA_UNICODE_STRING
	HighestRecordType mslsad.LSA_FOREST_TRUST_RECORD_TYPE `ndr:"enum"`
}

func (*lsarQueryForestTrustInformation2Request) Opnum() uint16 {
	return lsarpc.OpnumLsarQueryForestTrustInformation2
}

// lsarQueryForestTrustInformation2Response carries the [out] parameters and return value of LsarQueryForestTrustInformation2.
type lsarQueryForestTrustInformation2Response struct {
	ForestTrustInfo2 *mslsad.LSA_FOREST_TRUST_INFORMATION2 `ndr:"unique"`
	Status           ndr.DWORD                             `ndr:"retval"`
}

// LsarQueryForestTrustInformation2 calls LsarQueryForestTrustInformation2 (opnum 132) ([MS-LSAD]). Wire
// modeling is not yet live-validated against Windows.
func LsarQueryForestTrustInformation2(rpc ndr.Invoker, policyHandle mslsad.LSAPR_HANDLE, trustedDomainName mslsad.LSA_UNICODE_STRING, highestRecordType mslsad.LSA_FOREST_TRUST_RECORD_TYPE) (ForestTrustInfo2 *mslsad.LSA_FOREST_TRUST_INFORMATION2, err error) {
	req := &lsarQueryForestTrustInformation2Request{
		PolicyHandle:      policyHandle,
		TrustedDomainName: trustedDomainName,
		HighestRecordType: highestRecordType,
	}
	var resp lsarQueryForestTrustInformation2Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("LsarQueryForestTrustInformation2: %w", err)
		return
	}
	ForestTrustInfo2 = resp.ForestTrustInfo2
	if uint32(resp.Status) != lsarpc.StatusSuccess {
		err = fmt.Errorf("LsarQueryForestTrustInformation2 failed: %s", lsarpc.StatusString(uint32(resp.Status)))
	}
	return
}
