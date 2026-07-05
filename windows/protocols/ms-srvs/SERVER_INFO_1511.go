package mssrvs

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SERVER_INFO_1511 is a one-field SERVER_INFO level ([MS-SRVS] 2.2.4),
// carrying Sv1511Sessconns.
type SERVER_INFO_1511 struct {
	Sv1511Sessconns ndr.DWORD
}
