package mssrvs

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SERVER_INFO_1528 is a one-field SERVER_INFO level ([MS-SRVS] 2.2.4),
// carrying Sv1528Scavtimeout.
type SERVER_INFO_1528 struct {
	Sv1528Scavtimeout ndr.DWORD
}
