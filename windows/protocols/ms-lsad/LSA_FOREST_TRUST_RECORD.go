package mslsad

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// LSA_FOREST_TRUST_DATA is the discriminated union embedded as the ForestTrustData
// member of LSA_FOREST_TRUST_RECORD ([MS-LSAD] 2.2.7.21). The discriminant ForestTrustType
// is an LSA_FOREST_TRUST_RECORD_TYPE. The IDL maps both ForestTrustTopLevelName (0) and
// ForestTrustTopLevelNameEx (1) to the same TopLevelName arm; since the NDR walker
// matches a single case value per field, this is modeled as two distinct value fields of
// the same LSA_UNICODE_STRING (RPC_UNICODE_STRING) type, one per case value. The
// [default] arm carries an LSA_FOREST_TRUST_BINARY_DATA.
type LSA_FOREST_TRUST_DATA struct {
	ForestTrustType LSA_FOREST_TRUST_RECORD_TYPE `ndr:"switch,enum"`
	TopLevelName    msdtyp.RPC_UNICODE_STRING    `ndr:"case=0"`
	TopLevelNameEx  msdtyp.RPC_UNICODE_STRING    `ndr:"case=1"`
	DomainInfo      LSA_FOREST_TRUST_DOMAIN_INFO `ndr:"case=2"`
	Data            LSA_FOREST_TRUST_BINARY_DATA `ndr:"default"`
}

// LSA_FOREST_TRUST_RECORD is a single record of forest-trust information ([MS-LSAD]
// 2.2.7.21). ForestTrustData is a union switched on ForestTrustType.
type LSA_FOREST_TRUST_RECORD struct {
	Flags           ndr.DWORD
	ForestTrustType LSA_FOREST_TRUST_RECORD_TYPE `ndr:"enum"`
	Time            msdtyp.LARGE_INTEGER
	ForestTrustData LSA_FOREST_TRUST_DATA
}
