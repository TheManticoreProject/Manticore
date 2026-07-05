package mssrvs

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SERVER_INFO_1512 is a one-field SERVER_INFO level ([MS-SRVS] 2.2.4),
// carrying Sv1512Maxnonpagedmemoryusage.
type SERVER_INFO_1512 struct {
	Sv1512Maxnonpagedmemoryusage ndr.DWORD
}
