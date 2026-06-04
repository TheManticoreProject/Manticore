package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SERVER_INFO_1543 is a one-field SERVER_INFO level ([MS-SRVS] 2.2.4),
// carrying Sv1543Initsesstable.
type SERVER_INFO_1543 struct {
	Sv1543Initsesstable ndr.DWORD
}
