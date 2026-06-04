package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SERVER_INFO_1107 contains the count of users with simultaneous access to the
// server ([MS-SRVS] 2.2.4.48).
type SERVER_INFO_1107 struct {
	Sv1107Users ndr.DWORD
}
