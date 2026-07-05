package mssrvs

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SERVER_INFO_1546 is a one-field SERVER_INFO level ([MS-SRVS] 2.2.4),
// carrying Sv1546Initsearchtable.
type SERVER_INFO_1546 struct {
	Sv1546Initsearchtable ndr.DWORD
}
