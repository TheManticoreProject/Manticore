package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SERVER_INFO_1544 is a one-field SERVER_INFO level ([MS-SRVS] 2.2.4),
// carrying Sv1544Initconntable.
type SERVER_INFO_1544 struct {
	Sv1544Initconntable ndr.DWORD
}
