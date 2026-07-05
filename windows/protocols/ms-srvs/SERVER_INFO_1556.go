package mssrvs

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SERVER_INFO_1556 is a one-field SERVER_INFO level ([MS-SRVS] 2.2.4),
// carrying Sv1556Maxworkitemidletime.
type SERVER_INFO_1556 struct {
	Sv1556Maxworkitemidletime ndr.DWORD
}
