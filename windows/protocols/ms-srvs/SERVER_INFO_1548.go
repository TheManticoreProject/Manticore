package mssrvs

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SERVER_INFO_1548 is a one-field SERVER_INFO level ([MS-SRVS] 2.2.4),
// carrying Sv1548Errorthreshold.
type SERVER_INFO_1548 struct {
	Sv1548Errorthreshold ndr.DWORD
}
