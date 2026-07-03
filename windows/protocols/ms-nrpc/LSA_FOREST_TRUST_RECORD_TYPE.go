package msnrpc

// LSA_FOREST_TRUST_RECORD_TYPE is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-NRPC]).
type LSA_FOREST_TRUST_RECORD_TYPE uint16

const (
	ForestTrustTopLevelName   LSA_FOREST_TRUST_RECORD_TYPE = 0
	ForestTrustTopLevelNameEx LSA_FOREST_TRUST_RECORD_TYPE = 1
	ForestTrustDomainInfo     LSA_FOREST_TRUST_RECORD_TYPE = 2
	ForestTrustRecordTypeLast LSA_FOREST_TRUST_RECORD_TYPE = ForestTrustDomainInfo
)
