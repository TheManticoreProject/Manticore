package mswkst

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// USE_INFO is a discriminated union ([MS-WKST] 2.2.5.22). switch_type is unsigned long, so
// the discriminant is a DWORD, transmitted inline before the selected arm ([C706] 14.3.8).
// Every arm in the IDL is a [unique] pointer (LPUSE_INFO_0, ...), so each arm is a pointer
// field with `unique`; an inline-value arm would drop the arm referent id and desync decode.
type USE_INFO struct {
	Tag      ndr.DWORD   `ndr:"switch"`
	UseInfo0 *USE_INFO_0 `ndr:"case=0,unique"`
	UseInfo1 *USE_INFO_1 `ndr:"case=1,unique"`
	UseInfo2 *USE_INFO_2 `ndr:"case=2,unique"`
	UseInfo3 *USE_INFO_3 `ndr:"case=3,unique"`
}
