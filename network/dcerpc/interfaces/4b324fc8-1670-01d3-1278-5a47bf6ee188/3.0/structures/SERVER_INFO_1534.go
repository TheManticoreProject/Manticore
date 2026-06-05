package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SERVER_INFO_1534 is a one-field SERVER_INFO level ([MS-SRVS] 2.2.4),
// carrying Sv1534Oplockbreakwait.
type SERVER_INFO_1534 struct {
	Sv1534Oplockbreakwait ndr.DWORD
}
