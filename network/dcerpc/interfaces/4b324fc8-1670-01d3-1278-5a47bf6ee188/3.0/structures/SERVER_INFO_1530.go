package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SERVER_INFO_1530 is a one-field SERVER_INFO level ([MS-SRVS] 2.2.4),
// carrying Sv1530Minfreeworkitems.
type SERVER_INFO_1530 struct {
	Sv1530Minfreeworkitems ndr.DWORD
}
