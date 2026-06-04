package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SERVER_INFO_1549 is a one-field SERVER_INFO level ([MS-SRVS] 2.2.4),
// carrying Sv1549Networkerrorthreshold.
type SERVER_INFO_1549 struct {
	Sv1549Networkerrorthreshold ndr.DWORD
}
