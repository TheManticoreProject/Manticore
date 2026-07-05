package mssrvs

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SERVER_INFO_1503 is a one-field SERVER_INFO level ([MS-SRVS] 2.2.4),
// carrying Sv1503Opensearch.
type SERVER_INFO_1503 struct {
	Sv1503Opensearch ndr.DWORD
}
