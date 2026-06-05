package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SERVER_INFO_1523 is a one-field SERVER_INFO level ([MS-SRVS] 2.2.4),
// carrying Sv1523Maxkeepsearch.
type SERVER_INFO_1523 struct {
	Sv1523Maxkeepsearch ndr.DWORD
}
