package mssrvs

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SERVER_INFO_1502 is a one-field SERVER_INFO level ([MS-SRVS] 2.2.4),
// carrying Sv1502Sessvcs.
type SERVER_INFO_1502 struct {
	Sv1502Sessvcs ndr.DWORD
}
