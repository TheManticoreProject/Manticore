package mssrvs

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SERVER_INFO_1513 is a one-field SERVER_INFO level ([MS-SRVS] 2.2.4),
// carrying Sv1513Maxpagedmemoryusage.
type SERVER_INFO_1513 struct {
	Sv1513Maxpagedmemoryusage ndr.DWORD
}
