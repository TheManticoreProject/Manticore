package mssamr

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
)

// SAMPR_USER_HOME_INFORMATION holds a user's home directory and drive
// ([MS-SAMR] 2.2.6.22, HomeInformation).
type SAMPR_USER_HOME_INFORMATION struct {
	HomeDirectory      dtyp.RPC_UNICODE_STRING
	HomeDirectoryDrive dtyp.RPC_UNICODE_STRING
}
