package mssamr

import (
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// SAMPR_USER_PREFERENCES_INFORMATION holds user preference attributes
// ([MS-SAMR] 2.2.6.8).
type SAMPR_USER_PREFERENCES_INFORMATION struct {
	UserComment msdtyp.RPC_UNICODE_STRING
	Reserved1   msdtyp.RPC_UNICODE_STRING
	CountryCode uint16
	CodePage    uint16
}
