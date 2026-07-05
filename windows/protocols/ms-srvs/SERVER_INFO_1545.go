package mssrvs

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SERVER_INFO_1545 is a one-field SERVER_INFO level ([MS-SRVS] 2.2.4),
// carrying Sv1545Initfiletable.
type SERVER_INFO_1545 struct {
	Sv1545Initfiletable ndr.DWORD
}
