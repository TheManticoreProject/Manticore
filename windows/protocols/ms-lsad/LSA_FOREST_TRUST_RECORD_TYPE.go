package mslsad

// LSA_FOREST_TRUST_RECORD_TYPE enumerates the kinds of forest-trust record that select
// the ForestTrustData arm of LSA_FOREST_TRUST_RECORD ([MS-LSAD] 2.2.7.21). As an NDR
// enum it is transmitted as a 16-bit unsigned value ([C706] section 14.3.6).
type LSA_FOREST_TRUST_RECORD_TYPE uint16

const (
	ForestTrustTopLevelName   LSA_FOREST_TRUST_RECORD_TYPE = 0
	ForestTrustTopLevelNameEx LSA_FOREST_TRUST_RECORD_TYPE = 1
	ForestTrustDomainInfo     LSA_FOREST_TRUST_RECORD_TYPE = 2
	// ForestTrustBinaryInfo (3) and ForestTrustScannerInfo (4) are additional record
	// types used by the LSA_FOREST_TRUST_RECORD2 union ([MS-LSAD] 2.2.7.22a/2.2.7.29).
	ForestTrustBinaryInfo  LSA_FOREST_TRUST_RECORD_TYPE = 3
	ForestTrustScannerInfo LSA_FOREST_TRUST_RECORD_TYPE = 4
)
