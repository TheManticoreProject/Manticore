package mssrvs

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SHARE_INFO is the [switch_type(unsigned long)] union used by the share Get/Set info
// methods ([MS-SRVS] 2.2.3.6). The discriminant Tag is transmitted inline as a 4-byte
// DWORD followed by the selected arm, which is a [unique] pointer to the matching
// share-info structure. The IDL's empty [default]; arm is omitted: an unmatched
// discriminant decodes to an empty union.
type SHARE_INFO struct {
	Tag           ndr.DWORD          `ndr:"switch"`
	ShareInfo0    *SHARE_INFO_0      `ndr:"case=0,unique"`
	ShareInfo1    *SHARE_INFO_1      `ndr:"case=1,unique"`
	ShareInfo2    *SHARE_INFO_2      `ndr:"case=2,unique"`
	ShareInfo501  *SHARE_INFO_501    `ndr:"case=501,unique"`
	ShareInfo502  *SHARE_INFO_502_I  `ndr:"case=502,unique"`
	ShareInfo503  *SHARE_INFO_503_I  `ndr:"case=503,unique"`
	ShareInfo1004 *SHARE_INFO_1004   `ndr:"case=1004,unique"`
	ShareInfo1005 *SHARE_INFO_1005   `ndr:"case=1005,unique"`
	ShareInfo1006 *SHARE_INFO_1006   `ndr:"case=1006,unique"`
	ShareInfo1501 *SHARE_INFO_1501_I `ndr:"case=1501,unique"`
}
