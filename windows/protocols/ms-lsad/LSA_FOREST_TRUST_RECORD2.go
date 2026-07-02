package mslsad

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// LSA_FOREST_TRUST_DATA2 is the discriminated union embedded as the ForestTrustData
// member of LSA_FOREST_TRUST_RECORD2 ([MS-LSAD] 2.2.7.29). The discriminant
// ForestTrustType is an LSA_FOREST_TRUST_RECORD_TYPE. As with the v1 record, both
// ForestTrustTopLevelName (0) and ForestTrustTopLevelNameEx (1) select the same
// TopLevelName arm; since the NDR walker matches a single case value per field, this is
// modeled as two distinct value fields of the same LSA_UNICODE_STRING type, one per case
// value. Unlike the v1 record, the RECORD2 union declares explicit cases for
// ForestTrustBinaryInfo (3) and ForestTrustScannerInfo (4) rather than a [default] arm.
type LSA_FOREST_TRUST_DATA2 struct {
	ForestTrustType LSA_FOREST_TRUST_RECORD_TYPE  `ndr:"switch,enum"`
	TopLevelName    LSA_UNICODE_STRING            `ndr:"case=0"`
	TopLevelNameEx  LSA_UNICODE_STRING            `ndr:"case=1"`
	DomainInfo      LSA_FOREST_TRUST_DOMAIN_INFO  `ndr:"case=2"`
	BinaryData      LSA_FOREST_TRUST_BINARY_DATA  `ndr:"case=3"`
	ScannerInfo     LSA_FOREST_TRUST_SCANNER_INFO `ndr:"case=4"`
}

// LSA_FOREST_TRUST_RECORD2 is a single record of forest-trust information ([MS-LSAD]
// 2.2.7.29), the v2 counterpart of LSA_FOREST_TRUST_RECORD carrying the additional
// ForestTrustScannerInfo arm. ForestTrustData is a union switched on ForestTrustType.
type LSA_FOREST_TRUST_RECORD2 struct {
	Flags           ndr.DWORD
	ForestTrustType LSA_FOREST_TRUST_RECORD_TYPE `ndr:"enum"`
	Time            dtyp.LARGE_INTEGER
	ForestTrustData LSA_FOREST_TRUST_DATA2
}
