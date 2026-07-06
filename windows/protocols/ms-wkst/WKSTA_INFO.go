package mswkst

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// WKSTA_INFO is a discriminated union ([MS-WKST] 2.2.4.1). switch_type is unsigned long, so
// the discriminant is a DWORD, transmitted inline before the selected arm ([C706] 14.3.8).
// Every arm in the IDL is a [unique] pointer (LPWKSTA_INFO_100, ...), so each arm is a
// pointer field with `unique`: the codec emits the arm referent id after the discriminant
// and defers the body. Modeling an arm as an inline value would drop that referent id and
// desync the decode by four bytes.
type WKSTA_INFO struct {
	Tag           ndr.DWORD        `ndr:"switch"`
	WkstaInfo100  *WKSTA_INFO_100  `ndr:"case=100,unique"`
	WkstaInfo101  *WKSTA_INFO_101  `ndr:"case=101,unique"`
	WkstaInfo102  *WKSTA_INFO_102  `ndr:"case=102,unique"`
	WkstaInfo502  *WKSTA_INFO_502  `ndr:"case=502,unique"`
	WkstaInfo1013 *WKSTA_INFO_1013 `ndr:"case=1013,unique"`
	WkstaInfo1018 *WKSTA_INFO_1018 `ndr:"case=1018,unique"`
	WkstaInfo1046 *WKSTA_INFO_1046 `ndr:"case=1046,unique"`
}
