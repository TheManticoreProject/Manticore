package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SERVER_INFO_1506 is a one-field SERVER_INFO level ([MS-SRVS] 2.2.4),
// carrying Sv1506Maxworkitems.
type SERVER_INFO_1506 struct {
	Sv1506Maxworkitems ndr.DWORD
}
