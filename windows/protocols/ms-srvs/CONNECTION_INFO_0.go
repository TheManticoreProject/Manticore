package mssrvs

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// CONNECTION_INFO_0 contains the identifier of a connection ([MS-SRVS] 2.2.4.1).
type CONNECTION_INFO_0 struct {
	Coni0Id ndr.DWORD
}
