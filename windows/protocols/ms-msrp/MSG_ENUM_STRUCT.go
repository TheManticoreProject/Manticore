package msmsrp

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// MSG_ENUM_STRUCT is the [in,out] enumeration buffer of NetrMessageNameEnum
// ([MS-MSRP] 2.2.2.5). Level selects the MSG_ENUM_UNION arm; the union carries its own
// inline discriminant (Tag), which callers must set equal to Level before marshalling.
type MSG_ENUM_STRUCT struct {
	Level   ndr.DWORD
	MsgInfo MSG_ENUM_UNION
}

// MSG_ENUM_UNION is the [switch_is(Level)] union embedded in MSG_ENUM_STRUCT
// ([MS-MSRP] 2.2.2). The discriminant precedes the selected arm ([C706] 14.3.8). The
// arms are LPMSG_INFO_*_CONTAINER in the IDL (pointers to the container under
// pointer_default(unique)), so each is modeled as a unique pointer.
type MSG_ENUM_UNION struct {
	Tag    ndr.DWORD             `ndr:"switch"`
	Level0 *MSG_INFO_0_CONTAINER `ndr:"case=0,unique"`
	Level1 *MSG_INFO_1_CONTAINER `ndr:"case=1,unique"`
}
