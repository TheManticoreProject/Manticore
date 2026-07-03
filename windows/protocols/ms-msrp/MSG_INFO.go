package msmsrp

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// MSG_INFO is the [switch_type(DWORD)] union returned by NetrMessageNameGetInfo
// ([MS-MSRP] 2.2.2). The discriminant precedes the selected arm ([C706] 14.3.8). The
// arms are LPMSG_INFO_0/LPMSG_INFO_1 in the IDL (pointers under
// pointer_default(unique)), so each is modeled as a unique pointer.
type MSG_INFO struct {
	Tag      ndr.DWORD   `ndr:"switch"`
	MsgInfo0 *MSG_INFO_0 `ndr:"case=0,unique"`
	MsgInfo1 *MSG_INFO_1 `ndr:"case=1,unique"`
}
