package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SERVER_INFO_1017 contains the server announce rate ([MS-SRVS] 2.2.4.51).
type SERVER_INFO_1017 struct {
	Sv1017Announce ndr.DWORD
}
