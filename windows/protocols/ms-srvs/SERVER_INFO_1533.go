package mssrvs

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SERVER_INFO_1533 is a one-field SERVER_INFO level ([MS-SRVS] 2.2.4),
// carrying Sv1533Maxmpxct.
type SERVER_INFO_1533 struct {
	Sv1533Maxmpxct ndr.DWORD
}
