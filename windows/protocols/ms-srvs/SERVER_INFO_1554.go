package mssrvs

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SERVER_INFO_1554 is a one-field SERVER_INFO level ([MS-SRVS] 2.2.4),
// carrying Sv1554Linkinfovalidtime.
type SERVER_INFO_1554 struct {
	Sv1554Linkinfovalidtime ndr.DWORD
}
