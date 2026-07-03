package msmsrp

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// MSG_INFO_1 holds a message alias name plus its message-forwarding configuration
// ([MS-MSRP] 2.2.2.2). The name/forward fields are [string] wchar_t* (unique).
type MSG_INFO_1 struct {
	Msgi1_name         *ndr.WSTR `ndr:"unique"`
	Msgi1_forward_flag ndr.DWORD
	Msgi1_forward      *ndr.WSTR `ndr:"unique"`
}
