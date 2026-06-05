package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SERVER_INFO_1553 is a one-field SERVER_INFO level ([MS-SRVS] 2.2.4),
// carrying Sv1553Minlinkthroughput.
type SERVER_INFO_1553 struct {
	Sv1553Minlinkthroughput ndr.DWORD
}
