package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SERVER_INFO_1510 is a one-field SERVER_INFO level ([MS-SRVS] 2.2.4),
// carrying Sv1510Sessusers.
type SERVER_INFO_1510 struct {
	Sv1510Sessusers ndr.DWORD
}
