package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SERVER_INFO_1535 is a one-field SERVER_INFO level ([MS-SRVS] 2.2.4),
// carrying Sv1535Oplockbreakresponsewait.
type SERVER_INFO_1535 struct {
	Sv1535Oplockbreakresponsewait ndr.DWORD
}
