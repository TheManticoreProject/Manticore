package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SERVER_INFO_1552 is a one-field SERVER_INFO level ([MS-SRVS] 2.2.4),
// carrying Sv1552Maxlinkdelay.
type SERVER_INFO_1552 struct {
	Sv1552Maxlinkdelay ndr.DWORD
}
