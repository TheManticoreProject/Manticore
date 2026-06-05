package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SERVER_INFO_1550 is a one-field SERVER_INFO level ([MS-SRVS] 2.2.4),
// carrying Sv1550Diskspacethreshold.
type SERVER_INFO_1550 struct {
	Sv1550Diskspacethreshold ndr.DWORD
}
