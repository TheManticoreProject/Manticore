package mssrvs

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SERVER_INFO_1555 is a one-field SERVER_INFO level ([MS-SRVS] 2.2.4),
// carrying Sv1555Scavqosinfoupdatetime.
type SERVER_INFO_1555 struct {
	Sv1555Scavqosinfoupdatetime ndr.DWORD
}
