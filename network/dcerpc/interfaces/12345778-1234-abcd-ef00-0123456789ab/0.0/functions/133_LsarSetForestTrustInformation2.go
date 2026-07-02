package functions

import (
	"fmt"

	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mslsad "github.com/TheManticoreProject/Manticore/windows/protocols/ms-lsad"
)

// lsarSetForestTrustInformation2Request carries the [in] parameters of LsarSetForestTrustInformation2.
type lsarSetForestTrustInformation2Request struct {
	PolicyHandle      mslsad.LSAPR_HANDLE
	TrustedDomainName mslsad.LSA_UNICODE_STRING
	HighestRecordType mslsad.LSA_FOREST_TRUST_RECORD_TYPE `ndr:"enum"`
	ForestTrustInfo2  mslsad.LSA_FOREST_TRUST_INFORMATION2
	CheckOnly         uint8
}

func (*lsarSetForestTrustInformation2Request) Opnum() uint16 {
	return lsarpc.OpnumLsarSetForestTrustInformation2
}

// lsarSetForestTrustInformation2Response carries the [out] parameters and return value of LsarSetForestTrustInformation2.
type lsarSetForestTrustInformation2Response struct {
	CollisionInfo *mslsad.LSA_FOREST_TRUST_COLLISION_INFORMATION `ndr:"unique"`
	Status        ndr.DWORD                                      `ndr:"retval"`
}

// LsarSetForestTrustInformation2 calls LsarSetForestTrustInformation2 (opnum 133) ([MS-LSAD]). Wire
// modeling is not yet live-validated against Windows.
func LsarSetForestTrustInformation2(rpc ndr.Invoker, policyHandle mslsad.LSAPR_HANDLE, trustedDomainName mslsad.LSA_UNICODE_STRING, highestRecordType mslsad.LSA_FOREST_TRUST_RECORD_TYPE, forestTrustInfo2 mslsad.LSA_FOREST_TRUST_INFORMATION2, checkOnly uint8) (CollisionInfo *mslsad.LSA_FOREST_TRUST_COLLISION_INFORMATION, err error) {
	req := &lsarSetForestTrustInformation2Request{
		PolicyHandle:      policyHandle,
		TrustedDomainName: trustedDomainName,
		HighestRecordType: highestRecordType,
		ForestTrustInfo2:  forestTrustInfo2,
		CheckOnly:         checkOnly,
	}
	var resp lsarSetForestTrustInformation2Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("LsarSetForestTrustInformation2: %w", err)
		return
	}
	CollisionInfo = resp.CollisionInfo
	if uint32(resp.Status) != lsarpc.StatusSuccess {
		err = fmt.Errorf("LsarSetForestTrustInformation2 failed: %s", lsarpc.StatusString(uint32(resp.Status)))
	}
	return
}
