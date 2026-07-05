package mssrvs

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SHARE_ENUM_UNION is the [switch_type(DWORD)] union selecting a share-info container
// by enumeration level ([MS-SRVS] 2.2.4.38). The discriminant Tag is transmitted inline
// as a 4-byte DWORD followed by the selected arm, which is a [unique] pointer to the
// matching container.
type SHARE_ENUM_UNION struct {
	Tag      ndr.DWORD                 `ndr:"switch"`
	Level0   *SHARE_INFO_0_CONTAINER   `ndr:"case=0,unique"`
	Level1   *SHARE_INFO_1_CONTAINER   `ndr:"case=1,unique"`
	Level2   *SHARE_INFO_2_CONTAINER   `ndr:"case=2,unique"`
	Level501 *SHARE_INFO_501_CONTAINER `ndr:"case=501,unique"`
	Level502 *SHARE_INFO_502_CONTAINER `ndr:"case=502,unique"`
	Level503 *SHARE_INFO_503_CONTAINER `ndr:"case=503,unique"`
}
