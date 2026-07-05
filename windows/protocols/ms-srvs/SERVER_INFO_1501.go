package mssrvs

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SERVER_INFO_1501 is a one-field SERVER_INFO level ([MS-SRVS] 2.2.4),
// carrying Sv1501Sessopens.
type SERVER_INFO_1501 struct {
	Sv1501Sessopens ndr.DWORD
}
