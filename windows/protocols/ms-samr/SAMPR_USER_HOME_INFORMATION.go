package mssamr

import (
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// SAMPR_USER_HOME_INFORMATION holds a user's home directory and drive
// ([MS-SAMR] 2.2.6.22, HomeInformation).
type SAMPR_USER_HOME_INFORMATION struct {
	HomeDirectory      msdtyp.RPC_UNICODE_STRING
	HomeDirectoryDrive msdtyp.RPC_UNICODE_STRING
}
