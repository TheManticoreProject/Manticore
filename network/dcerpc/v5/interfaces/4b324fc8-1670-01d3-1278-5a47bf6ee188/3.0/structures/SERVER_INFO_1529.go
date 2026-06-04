package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SERVER_INFO_1529 is a one-field SERVER_INFO level ([MS-SRVS] 2.2.4),
// carrying Sv1529Minrcvqueue.
type SERVER_INFO_1529 struct {
	Sv1529Minrcvqueue ndr.DWORD
}
