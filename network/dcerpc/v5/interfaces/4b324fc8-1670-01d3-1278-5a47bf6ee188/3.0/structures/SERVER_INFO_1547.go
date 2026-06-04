package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SERVER_INFO_1547 is a one-field SERVER_INFO level ([MS-SRVS] 2.2.4),
// carrying Sv1547Alertschedule.
type SERVER_INFO_1547 struct {
	Sv1547Alertschedule ndr.DWORD
}
