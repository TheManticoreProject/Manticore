package mssrvs

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SERVER_INFO_1018 contains the server announce-rate delta ([MS-SRVS]
// 2.2.4.52).
type SERVER_INFO_1018 struct {
	Sv1018Anndelta ndr.DWORD
}
