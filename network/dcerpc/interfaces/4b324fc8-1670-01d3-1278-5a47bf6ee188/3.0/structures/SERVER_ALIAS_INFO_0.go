package structures

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// SERVER_ALIAS_INFO_0 contains the alias name and target server name
// ([MS-SRVS] 2.2.4.104). Srvai0Alias and Srvai0Target are [string] LMSTR
// (pointer_default unique).
type SERVER_ALIAS_INFO_0 struct {
	Srvai0Alias    ndr.WSTR `ndr:"unique"`
	Srvai0Target   ndr.WSTR `ndr:"unique"`
	Srvai0Default  bool
	Srvai0Reserved ndr.DWORD
}
